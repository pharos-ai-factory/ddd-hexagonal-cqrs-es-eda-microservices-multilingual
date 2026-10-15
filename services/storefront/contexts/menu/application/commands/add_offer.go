package commands

import (
	"context"
	"fmt"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// AddOfferCommandHandler applies AddOffer through one aggregate command transaction.
type AddOfferCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeEvent]
	Repository ports.EditionWriteRepository
	Drinks     a.ProjectionPort[app.DrinkPublishedDomainEvent]
}

func (h AddOfferCommandHandler) Execute(ctx context.Context, m a.CommandContext, c menucontracts.AddOfferCommand) (a.CommandResult, error) {
	// This immutable owner-local projection is loaded before the command transaction.
	published, found, err := h.Drinks.Find(ctx, fmt.Sprintf("%s/%d", c.DrinkID, c.DrinkRevision))
	if err != nil {
		return a.CommandResult{}, err
	}
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, core.Reject("not_found", "The edition does not exist")
	}
	if !found {
		return a.CommandResult{}, core.Reject("drink_revision_pending", "The published drink revision has not arrived")
	}
	edition := s.State
	offer, err := d.NewOffer(c.Code, d.DrinkState{ID: published.DrinkID, Name: published.Name, Revision: published.Revision, Published: true}, c.Minor, s.State.Snapshot().Currency)
	if err != nil {
		return a.CommandResult{}, err
	}
	if err = edition.AddOffer(offer); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Repository.Save(ctx, edition); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Realtime.Publish(ctx, a.RealtimeEvent{Snapshot: edition.Snapshot()}); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("draft"), nil
}
