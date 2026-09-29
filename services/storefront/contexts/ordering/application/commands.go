package application

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

type CreateOrder struct {
	CustomerID string `json:"customerId"`
	EditionID  string `json:"editionId"`
}
type CreateOrderHandler struct {
	Orders a.CommandPort[d.State]
	Menus  a.ProjectionPort[model.MenuPublished]
}

func (h CreateOrderHandler) Execute(ctx context.Context, m a.Metadata, c CreateOrder) (a.Outcome, error) {
	menu, found, err := h.Menus.Find(ctx, c.EditionID)
	if err != nil {
		return a.Outcome{}, err
	}
	return h.Orders.Execute(ctx, m, func(s a.Loaded[d.State]) (a.Mutation[d.State], error) {
		if s.Exists {
			return a.Mutation[d.State]{}, core.Reject("already_exists", "The order already exists")
		}
		if !found {
			return a.Mutation[d.State]{}, core.Reject("menu_pending", "The published edition has not arrived")
		}
		order, err := d.New(m.AggregateID, c.CustomerID, c.EditionID, menu.Currency)
		if err != nil {
			return a.Mutation[d.State]{}, err
		}
		return a.Changed(order.Snapshot(), "draft"), nil
	})
}

type AddLine struct {
	LineID    string `json:"lineId"`
	EditionID string `json:"editionId"`
	OfferCode string `json:"offerCode"`
	Quantity  int    `json:"quantity"`
}
type AddLineHandler struct {
	Orders a.CommandPort[d.State]
	Menus  a.ProjectionPort[model.MenuPublished]
}

func (h AddLineHandler) Execute(ctx context.Context, m a.Metadata, c AddLine) (a.Outcome, error) {
	menu, found, err := h.Menus.Find(ctx, c.EditionID)
	if err != nil {
		return a.Outcome{}, err
	}
	return h.Orders.Execute(ctx, m, func(s a.Loaded[d.State]) (a.Mutation[d.State], error) {
		if !s.Exists {
			return a.Mutation[d.State]{}, orderNotFound()
		}
		if !found || s.State.EditionID != c.EditionID {
			return a.Mutation[d.State]{}, core.Reject("incorrect_edition", "The selection must belong to the order's edition")
		}
		order, err := d.Restore(s.State)
		if err != nil {
			return a.Mutation[d.State]{}, err
		}
		for _, offer := range menu.Offers {
			if offer.Code == c.OfferCode {
				if err = order.AddLine(c.LineID, d.Selection{OfferCode: offer.Code, Name: offer.Name, Minor: offer.Minor}, c.Quantity); err != nil {
					return a.Mutation[d.State]{}, err
				}
				return a.Changed(order.Snapshot(), "draft"), nil
			}
		}
		return a.Mutation[d.State]{}, core.Reject("offer_not_found", "The published menu has no such offer")
	})
}

type ChangeQuantity struct {
	LineID   string `json:"lineId"`
	Quantity int    `json:"quantity"`
}
type ChangeQuantityHandler struct{ Orders a.CommandPort[d.State] }

func (h ChangeQuantityHandler) Execute(ctx context.Context, m a.Metadata, c ChangeQuantity) (a.Outcome, error) {
	return h.Orders.Execute(ctx, m, func(s a.Loaded[d.State]) (a.Mutation[d.State], error) {
		if !s.Exists {
			return a.Mutation[d.State]{}, orderNotFound()
		}
		order, err := d.Restore(s.State)
		if err != nil {
			return a.Mutation[d.State]{}, err
		}
		if err = order.ChangeQuantity(c.LineID, c.Quantity); err != nil {
			return a.Mutation[d.State]{}, err
		}
		return a.Mutation[d.State]{State: order.Snapshot(), Changed: len(order.Events()) > 0, Status: "draft"}, nil
	})
}

type PlaceOrderHandler struct{ Orders a.CommandPort[d.State] }

func (h PlaceOrderHandler) Execute(ctx context.Context, m a.Metadata, _ struct{}) (a.Outcome, error) {
	return h.Orders.Execute(ctx, m, func(s a.Loaded[d.State]) (a.Mutation[d.State], error) {
		if !s.Exists {
			return a.Mutation[d.State]{}, orderNotFound()
		}
		order, err := d.Restore(s.State)
		if err != nil {
			return a.Mutation[d.State]{}, err
		}
		if err = order.Place(); err != nil {
			return a.Mutation[d.State]{}, err
		}
		payload := model.OrderPlaced{OrderID: s.State.ID, CustomerID: s.State.CustomerID, EditionID: s.State.EditionID, Currency: s.State.Currency, Lines: []model.Line{}}
		for _, line := range order.Snapshot().Lines {
			payload.Lines = append(payload.Lines, model.Line{ID: line.ID, OfferCode: line.Selection.OfferCode, Name: line.Selection.Name, Quantity: line.Quantity, Minor: line.Selection.Minor})
		}
		return a.Changed(order.Snapshot(), "placed", a.Publication{Name: "ordering.order-placed", Visibility: a.Public, Payload: payload}), nil
	})
}
