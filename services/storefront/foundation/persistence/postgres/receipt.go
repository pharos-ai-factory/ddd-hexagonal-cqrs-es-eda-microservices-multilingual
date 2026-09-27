package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Receipt ownership is serialised independently of the target aggregate.
// Concurrent conflicting copies of one event cannot affect different targets.
func loadIncoming(ctx context.Context, tx pgx.Tx, m a.Metadata, target string) (bool, a.Outcome, error) {
	if m.Consumer == "" {
		return false, a.Outcome{}, nil
	}
	key := "consumer:" + m.Consumer + ":" + m.SourceEventID
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, key); err != nil {
		return false, a.Outcome{}, err
	}
	var hash, originalTarget string
	var data []byte
	err := tx.QueryRow(ctx, `SELECT fingerprint,target,outcome FROM cafe.consumer_receipts WHERE consumer=$1 AND event_id=$2`,
		m.Consumer, m.SourceEventID).Scan(&hash, &originalTarget, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, a.Outcome{}, nil
	}
	if err != nil {
		return false, a.Outcome{}, err
	}
	if hash != m.SourceHash || target != originalTarget {
		return false, a.Outcome{}, fmt.Errorf("event identity reused with conflicting bytes or target")
	}
	var outcome a.Outcome
	if err = json.Unmarshal(data, &outcome); err != nil {
		return false, a.Outcome{}, err
	}
	return true, outcome, nil
}

func recordIncoming(ctx context.Context, tx pgx.Tx, m a.Metadata, target string, outcome a.Outcome) error {
	if m.Consumer == "" {
		return nil
	}
	data, err := json.Marshal(outcome)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome)
		VALUES($1,$2,$3,$4,$5)`, m.Consumer, m.SourceEventID, m.SourceHash, target, data)
	return err
}
