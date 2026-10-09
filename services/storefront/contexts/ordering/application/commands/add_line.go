package commands

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

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
