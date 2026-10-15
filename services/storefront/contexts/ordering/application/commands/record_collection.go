package commands

import (
	"context"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// RecordCollectionCommand records Collection's published handover fact.
type RecordCollectionCommand struct {
	CustomerID string `json:"customerId"`
}

// RecordCollectionCommandHandler changes one Order and explicitly publishes its updated view.
type RecordCollectionCommandHandler struct {
	Repository ports.OrderWriteRepository
	Realtime   a.RealtimeBusPort[a.RealtimeEvent]
}

func (h RecordCollectionCommandHandler) Execute(ctx context.Context, m a.CommandContext, c RecordCollectionCommand) (a.CommandResult, error) {
	loaded, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !loaded.Exists {
		return a.CommandResult{}, fmt.Errorf("collected order is missing")
	}
	order := loaded.State
	if err = order.RecordCollection(c.CustomerID); err != nil {
		return a.CommandResult{}, err
	}
	if len(order.Events()) > 0 {
		if err = h.Repository.Save(ctx, order); err != nil {
			return a.CommandResult{}, err
		}
		if err = h.Realtime.Publish(ctx, a.RealtimeEvent{Snapshot: order.Snapshot()}); err != nil {
			return a.CommandResult{}, err
		}
	}
	return a.Result("collection_recorded"), nil
}
