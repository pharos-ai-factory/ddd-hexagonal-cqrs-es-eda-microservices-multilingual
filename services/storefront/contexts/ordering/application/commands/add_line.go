package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// AddLineCommandHandler applies AddLine through one aggregate command transaction.
type AddLineCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeEvent]
	Repository ports.OrderWriteRepository
	Menus      a.ProjectionPort[model.MenuPublishedIntegrationEvent]
}

func (h AddLineCommandHandler) Execute(ctx context.Context, m a.CommandContext, c orderingcontracts.AddLineCommand) (a.CommandResult, error) {
	menu, found, err := h.Menus.Find(ctx, c.EditionID)
	if err != nil {
		return a.CommandResult{}, err
	}
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, orderNotFound()
	}
	if !found || s.State.Snapshot().EditionID != c.EditionID {
		return a.CommandResult{}, core.Reject("incorrect_edition", "The selection must belong to the order's edition")
	}
	order := s.State
	for _, offer := range menu.Offers {
		if offer.Code == c.OfferCode {
			if err = order.AddLine(c.LineID, d.Selection{OfferCode: offer.Code, Name: offer.Name, Minor: offer.Minor}, c.Quantity); err != nil {
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
	}
	return a.CommandResult{}, core.Reject("offer_not_found", "The published menu has no such offer")
}
