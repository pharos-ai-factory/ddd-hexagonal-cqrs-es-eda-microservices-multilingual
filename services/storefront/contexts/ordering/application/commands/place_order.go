package commands

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// PlaceOrderCommand expresses the owner use case independently of its transport.
type PlaceOrderCommand struct{}

// PlaceOrderCommandHandler applies PlaceOrder through one aggregate command transaction.
type PlaceOrderCommandHandler struct {
	Orders a.AggregateCommandPort[d.OrderState]
}

func (h PlaceOrderCommandHandler) Execute(ctx context.Context, m a.Metadata, _ PlaceOrderCommand) (a.Outcome, error) {
	return h.Orders.Execute(ctx, m, func(s a.Loaded[d.OrderState]) (a.Mutation[d.OrderState], error) {
		if !s.Exists {
			return a.Mutation[d.OrderState]{}, orderNotFound()
		}
		order, err := d.Restore(s.State)
		if err != nil {
			return a.Mutation[d.OrderState]{}, err
		}
		if err = order.Place(); err != nil {
			return a.Mutation[d.OrderState]{}, err
		}
		payload := model.OrderPlaced{OrderID: s.State.ID, CustomerID: s.State.CustomerID, EditionID: s.State.EditionID, Currency: s.State.Currency, Lines: []model.Line{}}
		for _, line := range order.Snapshot().Lines {
			payload.Lines = append(payload.Lines, model.Line{ID: line.ID, OfferCode: line.Selection.OfferCode, Name: line.Selection.Name, Quantity: line.Quantity, Minor: line.Selection.Minor})
		}
		return a.Changed(order.Snapshot(), "placed", a.Publication{Name: "ordering.order-placed", Visibility: a.Public, Payload: payload}), nil
	})
}
