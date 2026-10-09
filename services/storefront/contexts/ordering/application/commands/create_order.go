package commands

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// CreateOrderCommand expresses the owner CreateOrder use case independently of transport.
type CreateOrderCommand struct {
	CustomerID string `json:"customerId"`
	EditionID  string `json:"editionId"`
}

// CreateOrderCommandHandler applies CreateOrder through one aggregate command transaction.
type CreateOrderCommandHandler struct {
	Orders a.AggregateCommandPort[d.OrderState]
	Menus  a.ProjectionPort[model.MenuPublished]
}

func (h CreateOrderCommandHandler) Execute(ctx context.Context, m a.Metadata, c CreateOrderCommand) (a.Outcome, error) {
	menu, found, err := h.Menus.Find(ctx, c.EditionID)
	if err != nil {
		return a.Outcome{}, err
	}
	return h.Orders.Execute(ctx, m, func(s a.Loaded[d.OrderState]) (a.Mutation[d.OrderState], error) {
		if s.Exists {
			return a.Mutation[d.OrderState]{}, core.Reject("already_exists", "The order already exists")
		}
		if !found {
			return a.Mutation[d.OrderState]{}, core.Reject("menu_pending", "The published edition has not arrived")
		}
		order, err := d.New(m.AggregateID, c.CustomerID, c.EditionID, menu.Currency)
		if err != nil {
			return a.Mutation[d.OrderState]{}, err
		}
		return a.Changed(order.Snapshot(), "draft"), nil
	})
}
