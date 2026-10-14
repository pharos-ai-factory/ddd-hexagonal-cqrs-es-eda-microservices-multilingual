package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

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
