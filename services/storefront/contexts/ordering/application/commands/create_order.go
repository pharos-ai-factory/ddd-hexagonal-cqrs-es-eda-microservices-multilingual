package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/model"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// CreateOrderCommandHandler applies CreateOrder through one aggregate command transaction.
type CreateOrderCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeEvent]
	Repository ports.OrderWriteRepository
	Menus      a.ProjectionPort[model.MenuPublishedIntegrationEvent]
}

func (h CreateOrderCommandHandler) Execute(ctx context.Context, m a.CommandContext, c orderingcontracts.CreateOrderCommand) (a.CommandResult, error) {
	menu, found, err := h.Menus.Find(ctx, c.EditionID)
	if err != nil {
		return a.CommandResult{}, err
	}
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if s.Exists {
		return a.CommandResult{}, core.Reject("already_exists", "The order already exists")
	}
	if !found {
		return a.CommandResult{}, core.Reject("menu_pending", "The published edition has not arrived")
	}
	order, err := d.New(m.Target, c.CustomerID, c.EditionID, menu.Currency)
	if err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Repository.Save(ctx, order); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Realtime.Publish(ctx, a.RealtimeEvent{Snapshot: order.Snapshot()}); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("draft"), nil
}
