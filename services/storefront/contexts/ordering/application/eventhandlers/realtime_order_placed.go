package eventhandlers

import (
	"context"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderPlacedRealtimeDomainEventHandler publishes the current owner view with its actual revision.
type OrderPlacedRealtimeDomainEventHandler struct {
	Orders   ports.OrderReadRepository
	Realtime a.RealtimeBusPort[a.RealtimeMessage]
}

func (h OrderPlacedRealtimeDomainEventHandler) Handle(ctx context.Context, m a.Metadata, event d.OrderPlacedDomainEvent) (a.Outcome, error) {
	if event.State.ID != m.AggregateID {
		return a.Outcome{}, fmt.Errorf("order event target mismatch")
	}
	loaded, err := h.Orders.Get(ctx, m.AggregateID)
	if err != nil {
		return a.Outcome{}, err
	}
	if !loaded.Exists {
		return a.Outcome{}, fmt.Errorf("committed order is missing")
	}
	message := a.SnapshotMessage(a.ContextTarget("ordering"), loaded.State)
	message.Source, message.Revision = &m, loaded.Version
	err = h.Realtime.Publish(ctx, message)
	return a.Outcome{AggregateID: m.AggregateID, Version: loaded.Version, Status: "realtime_queued"}, err
}
