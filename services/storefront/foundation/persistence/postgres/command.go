package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"time"
)

// AggregateCommandStore implements atomic aggregate decisions, receipts, outcomes and outgoing intent.
type AggregateCommandStore[S any] struct {
	db   *ContextDatabase
	kind string
}

func Command[S any](db *ContextDatabase, kind string) *AggregateCommandStore[S] {
	return &AggregateCommandStore[S]{db: db, kind: kind}
}
func (s *AggregateCommandStore[S]) Execute(ctx context.Context, m a.Metadata, decide func(a.Loaded[S]) (a.Mutation[S], error)) (a.Outcome, error) {
	for _, id := range []string{m.ID, m.AggregateID, m.CorrelationID} {
		if err := d.ValidateID(id); err != nil {
			return a.Outcome{}, err
		}
	}
	hash, err := fingerprint(struct {
		Expected *uint64
		Input    any
	}{m.ExpectedVersion, m.Input})
	if err != nil {
		return a.Outcome{}, err
	}
	tx, err := begin(ctx, s.db.pool)
	if err != nil {
		return a.Outcome{}, err
	}
	defer tx.Rollback(context.Background())
	// Serialise creation and receipt ownership even before an aggregate row exists.
	target := s.kind + ":" + m.AggregateID
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, target); err != nil {
		return a.Outcome{}, err
	}
	if _, err = tx.Exec(ctx, `SELECT set_config('cafe.command_target',$1,true)`, target); err != nil {
		return a.Outcome{}, err
	}
	done, previous, err := loadIncoming(ctx, tx, m, target, m.AggregateID)
	if err != nil {
		return a.Outcome{}, err
	}
	if done {
		return commitOutcome(ctx, tx, previous)
	}
	var savedHash string
	var saved []byte
	err = tx.QueryRow(ctx, `SELECT fingerprint,outcome FROM cafe.command_receipts WHERE kind=$1 AND aggregate_id=$2 AND command_name=$3 AND command_id=$4`, s.kind, m.AggregateID, m.Name, m.ID).Scan(&savedHash, &saved)
	if err == nil {
		if savedHash != hash {
			detail := (&a.ApplicationError{Code: "idempotency_conflict", Message: "The command identity was reused with different input"}).Rejection()
			return commitOutcome(ctx, tx, a.Outcome{AggregateID: m.AggregateID, Rejection: &detail})
		}
		outcome, err := decodeOutcome(saved, m.AggregateID)
		if err != nil {
			return a.Outcome{}, err
		}
		if err = recordIncoming(ctx, tx, m, target, outcome); err != nil {
			return a.Outcome{}, err
		}
		return commitOutcome(ctx, tx, outcome)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return a.Outcome{}, err
	}
	var loaded a.Loaded[S]
	var state []byte
	err = tx.QueryRow(ctx, `SELECT version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2 FOR UPDATE`, s.kind, m.AggregateID).Scan(&loaded.Version, &state)
	if err == nil {
		loaded.Exists = true
		if err = decodeStored(state, &loaded.State); err != nil {
			return a.Outcome{}, err
		}
		if err = checkRootIdentity(state, m.AggregateID); err != nil {
			return a.Outcome{}, err
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return a.Outcome{}, err
	}
	outcome := a.Outcome{AggregateID: m.AggregateID, Version: loaded.Version}
	var mutation a.Mutation[S]
	if m.ExpectedVersion != nil && *m.ExpectedVersion != loaded.Version {
		detail := (&a.ApplicationError{Code: "version_conflict", Message: "The expected aggregate version is stale"}).Rejection()
		outcome.Rejection = &detail
	} else {
		mutation, err = decide(loaded)
		if err != nil {
			var expected d.ExpectedError
			if !errors.As(err, &expected) {
				return a.Outcome{}, err
			}
			detail := expected.Rejection()
			outcome.Rejection = &detail
		} else {
			outcome.Status = mutation.Status
			if err = s.persistMutation(ctx, tx, m, loaded, mutation, &outcome); err != nil {
				return a.Outcome{}, err
			}
		}
	}
	data, err := json.Marshal(outcome)
	if err != nil {
		return a.Outcome{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO cafe.command_receipts(kind,aggregate_id,command_name,command_id,fingerprint,outcome,correlation_id,causation_id) VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($8,'')::uuid)`, s.kind, m.AggregateID, m.Name, m.ID, hash, data, m.CorrelationID, m.CausationID); err != nil {
		return a.Outcome{}, err
	}
	if err = recordIncoming(ctx, tx, m, target, outcome); err != nil {
		return a.Outcome{}, err
	}
	return commitOutcome(ctx, tx, outcome)
}

// persistMutation writes the root and both publication intents on the command's transaction.
func (s *AggregateCommandStore[S]) persistMutation(ctx context.Context, tx pgx.Tx, m a.Metadata, loaded a.Loaded[S], mutation a.Mutation[S], outcome *a.Outcome) error {
	if !mutation.Changed {
		if len(mutation.Publications) > 0 {
			return fmt.Errorf("a no-op cannot publish new events")
		}
		return nil
	}
	state, err := json.Marshal(mutation.State)
	if err != nil {
		return err
	}
	if err = checkRootIdentity(state, m.AggregateID); err != nil {
		return err
	}
	outcome.Version++
	if loaded.Exists {
		tag, writeErr := tx.Exec(ctx, `UPDATE cafe.aggregates SET version=$3,state=$4 WHERE kind=$1 AND id=$2 AND version=$5`, s.kind, m.AggregateID, outcome.Version, state, loaded.Version)
		if writeErr != nil {
			return writeErr
		}
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("optimistic update lost its owner version")
		}
	} else if _, err = tx.Exec(ctx, `INSERT INTO cafe.aggregates(kind,id,version,state) VALUES($1,$2,$3,$4)`, s.kind, m.AggregateID, outcome.Version, state); err != nil {
		return err
	}
	for _, publication := range mutation.Publications {
		message := a.Message{ID: NewID(), Name: publication.Name, Context: s.db.owner, Visibility: publication.Visibility, ContractVersion: 1, AggregateKind: s.kind, AggregateID: m.AggregateID, AggregateVersion: outcome.Version, CorrelationID: m.CorrelationID, CausationID: m.ID, OccurredAt: time.Now().UTC(), Payload: publication.Payload}
		if err = s.db.appendEvent(ctx, tx, message); err != nil {
			return err
		}
	}
	if s.db.realtime != nil {
		return s.db.appendRealtime(ctx, tx, s.kind, m.AggregateID, outcome.Version, state)
	}
	return nil
}

func (db *ContextDatabase) appendEvent(ctx context.Context, tx pgx.Tx, message a.Message) error {
	encoded, err := db.encode(message)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO cafe.outbox_events(id,event_name,visibility,aggregate_kind,aggregate_id,aggregate_version,correlation_id,causation_id,body) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, message.ID, message.Name, string(message.Visibility), message.AggregateKind, message.AggregateID, message.AggregateVersion, message.CorrelationID, message.CausationID, encoded)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO cafe.dispatches(event_id) VALUES($1)`, message.ID)
	return err
}
