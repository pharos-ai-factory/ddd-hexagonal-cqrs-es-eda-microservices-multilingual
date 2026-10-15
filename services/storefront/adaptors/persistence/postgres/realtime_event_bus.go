package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"regexp"
	"strings"
)

// RealtimeEventBus commits one declared event subscription's receipt and exact browser bytes.
type RealtimeEventBus struct {
	db             *ContextDatabase
	consumer, kind string
}

func NewRealtimeEventBus(db *ContextDatabase, consumer, kind string) *RealtimeEventBus {
	if !strings.HasPrefix(consumer, db.owner+".") {
		panic("foreign realtime subscription")
	}
	return &RealtimeEventBus{db, consumer, kind}
}
func (b *RealtimeEventBus) Publish(ctx context.Context, event a.RealtimeEvent) error {
	if event.Source == nil || event.Snapshot == nil || event.Rejection != nil || event.Revision == 0 || event.Revision > 9007199254740991 {
		return fmt.Errorf("invalid realtime delivery")
	}
	m := *event.Source
	for _, id := range []string{m.ID, m.AggregateID, m.SourceEventID, m.CorrelationID} {
		if err := d.ValidateID(id); err != nil {
			return err
		}
	}
	if m.Consumer != b.consumer || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(m.SourceHash) {
		return fmt.Errorf("invalid realtime receipt")
	}
	data, err := json.Marshal(event.Snapshot)
	if err != nil {
		return err
	}
	if err = checkRootIdentity(data, m.AggregateID); err != nil {
		return err
	}
	tx, err := begin(ctx, b.db.pool)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	target := "realtime:" + b.kind + ":" + m.AggregateID
	done, _, err := loadIncoming(ctx, tx, m, target, m.AggregateID)
	if err != nil {
		return err
	}
	if done {
		return tx.Commit(ctx)
	}
	if err = b.db.appendRealtime(ctx, tx, b.kind, m.AggregateID, event.Revision, event); err != nil {
		return err
	}
	outcome := a.Outcome{AggregateID: m.AggregateID, Version: event.Revision, Status: "realtime_queued"}
	if err = recordIncoming(ctx, tx, m, target, outcome); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
