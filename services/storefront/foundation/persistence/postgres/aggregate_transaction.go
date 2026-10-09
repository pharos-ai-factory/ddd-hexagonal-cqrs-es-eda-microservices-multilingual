package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	"time"
)

// AggregateTransaction coordinates aggregate persistence, receipts, outcomes and outgoing intent.
type AggregateTransaction[S, A any] struct {
	db           *ContextDatabase
	kind         string
	repository   func(a.WriteRepository[S]) a.WriteRepository[A]
	publications func(A) []a.Publication
}

// NewAggregateTransaction constructs a transaction boundary with an owner write repository factory.
func NewAggregateTransaction[S, A any](db *ContextDatabase, kind string, repository func(a.WriteRepository[S]) a.WriteRepository[A], publications ...func(A) []a.Publication) *AggregateTransaction[S, A] {
	result := &AggregateTransaction[S, A]{db: db, kind: kind, repository: repository}
	if len(publications) > 0 {
		result.publications = publications[0]
	}
	return result
}
func (s *AggregateTransaction[S, A]) Execute(ctx context.Context, m a.Metadata, work func(a.WriteRepository[A]) (execution.Result, error)) (a.Outcome, error) {
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
	repository, err := newSnapshotWriteRepository[S](ctx, tx, s.kind, m.AggregateID)
	if err != nil {
		return a.Outcome{}, err
	}
	defer repository.close()
	loaded := repository.loaded
	outcome := a.Outcome{AggregateID: m.AggregateID, Version: loaded.Version}
	var result execution.Result
	if m.ExpectedVersion != nil && *m.ExpectedVersion != loaded.Version {
		detail := (&a.ApplicationError{Code: "version_conflict", Message: "The expected aggregate version is stale"}).Rejection()
		outcome.Rejection = &detail
	} else {
		ownerRepository := &execution.EventRecordingRepository[A]{Source: s.repository(repository), Map: s.publications}
		result, err = work(ownerRepository)
		result.Publications = append(result.Publications, ownerRepository.Publications...)
		repository.close()
		if err != nil {
			var expected d.ExpectedError
			if !errors.As(err, &expected) {
				return a.Outcome{}, err
			}
			detail := expected.Rejection()
			outcome.Rejection = &detail
		} else {
			outcome.Status = result.Status
			if err = s.persistResult(ctx, tx, m, repository, result, &outcome); err != nil {
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

// persistResult flushes the saved root and appends outgoing intent on the same transaction.
func (s *AggregateTransaction[S, A]) persistResult(ctx context.Context, tx pgx.Tx, m a.Metadata, repository *SnapshotWriteRepository[S], result execution.Result, outcome *a.Outcome) error {
	if repository.pending == nil {
		if len(result.Publications) > 0 {
			return fmt.Errorf("a no-op cannot publish new events")
		}
		return nil
	}
	version, state, err := repository.flush(ctx)
	if err != nil {
		return err
	}
	outcome.Version = version
	for _, publication := range result.Publications {
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
