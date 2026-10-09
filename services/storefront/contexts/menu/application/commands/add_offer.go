package commands

import (
	"context"
	"fmt"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// AddOfferCommand expresses the owner AddOffer use case independently of transport.
type AddOfferCommand struct {
	Code          string `json:"code"`
	DrinkID       string `json:"drinkId"`
	DrinkRevision uint64 `json:"drinkRevision"`
	Minor         int64  `json:"minor"`
}

// AddOfferCommandHandler applies AddOffer through one aggregate command transaction.
type AddOfferCommandHandler struct {
	Editions a.AggregateCommandPort[d.EditionState]
	Drinks   a.ProjectionPort[app.DrinkPublished]
}

func (h AddOfferCommandHandler) Execute(ctx context.Context, m a.Metadata, c AddOfferCommand) (a.Outcome, error) {
	// This immutable owner-local projection is loaded before the command transaction.
	published, found, err := h.Drinks.Find(ctx, fmt.Sprintf("%s/%d", c.DrinkID, c.DrinkRevision))
	if err != nil {
		return a.Outcome{}, err
	}
	return h.Editions.Execute(ctx, m, func(s a.Loaded[d.EditionState]) (a.Mutation[d.EditionState], error) {
		if !s.Exists {
			return a.Mutation[d.EditionState]{}, core.Reject("not_found", "The edition does not exist")
		}
		if !found {
			return a.Mutation[d.EditionState]{}, core.Reject("drink_revision_pending", "The published drink revision has not arrived")
		}
		edition, err := d.RestoreEdition(s.State)
		if err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		offer, err := d.NewOffer(c.Code, d.DrinkState{ID: published.DrinkID, Name: published.Name, Revision: published.Revision, Published: true}, c.Minor, s.State.Currency)
		if err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		if err = edition.AddOffer(offer); err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		return a.Changed(edition.Snapshot(), "draft"), nil
	})
}
