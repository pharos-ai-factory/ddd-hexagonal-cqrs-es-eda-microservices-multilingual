package commands

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ChangeQuantityCommand expresses the owner ChangeQuantity use case independently of transport.
type ChangeQuantityCommand struct {
	LineID   string `json:"lineId"`
	Quantity int    `json:"quantity"`
}

// ChangeQuantityCommandHandler applies ChangeQuantity through one aggregate command transaction.
type ChangeQuantityCommandHandler struct {
	Orders a.AggregateCommandPort[d.OrderState]
}

func (h ChangeQuantityCommandHandler) Execute(ctx context.Context, m a.Metadata, c ChangeQuantityCommand) (a.Outcome, error) {
	return h.Orders.Execute(ctx, m, func(s a.Loaded[d.OrderState]) (a.Mutation[d.OrderState], error) {
		if !s.Exists {
			return a.Mutation[d.OrderState]{}, orderNotFound()
		}
		order, err := d.Restore(s.State)
		if err != nil {
			return a.Mutation[d.OrderState]{}, err
		}
		if err = order.ChangeQuantity(c.LineID, c.Quantity); err != nil {
			return a.Mutation[d.OrderState]{}, err
		}
		return a.Mutation[d.OrderState]{State: order.Snapshot(), Changed: len(order.Events()) > 0, Status: "draft"}, nil
	})
}
