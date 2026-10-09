package application

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

// AddLineCommand expresses the owner AddLine use case independently of transport.
type AddLineCommand struct {
	LineID    string `json:"lineId"`
	EditionID string `json:"editionId"`
	OfferCode string `json:"offerCode"`
	Quantity  int    `json:"quantity"`
}

// AddLineCommandHandler applies AddLine through one aggregate command transaction.
type AddLineCommandHandler struct {
	Orders a.AggregateCommandPort[d.OrderState]
	Menus  a.ProjectionPort[model.MenuPublished]
}

func (h AddLineCommandHandler) Execute(ctx context.Context, m a.Metadata, c AddLineCommand) (a.Outcome, error) {
	menu, found, err := h.Menus.Find(ctx, c.EditionID)
	if err != nil {
		return a.Outcome{}, err
	}
	return h.Orders.Execute(ctx, m, func(s a.Loaded[d.OrderState]) (a.Mutation[d.OrderState], error) {
		if !s.Exists {
			return a.Mutation[d.OrderState]{}, orderNotFound()
		}
		if !found || s.State.EditionID != c.EditionID {
			return a.Mutation[d.OrderState]{}, core.Reject("incorrect_edition", "The selection must belong to the order's edition")
		}
		order, err := d.Restore(s.State)
		if err != nil {
			return a.Mutation[d.OrderState]{}, err
		}
		for _, offer := range menu.Offers {
			if offer.Code == c.OfferCode {
				if err = order.AddLine(c.LineID, d.Selection{OfferCode: offer.Code, Name: offer.Name, Minor: offer.Minor}, c.Quantity); err != nil {
					return a.Mutation[d.OrderState]{}, err
				}
				return a.Changed(order.Snapshot(), "draft"), nil
			}
		}
		return a.Mutation[d.OrderState]{}, core.Reject("offer_not_found", "The published menu has no such offer")
	})
}

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

// PlaceOrderCommand expresses the owner use case independently of its transport.
type PlaceOrderCommand struct{}
