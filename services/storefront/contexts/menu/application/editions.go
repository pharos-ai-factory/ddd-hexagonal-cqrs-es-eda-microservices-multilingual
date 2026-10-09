package application

import (
	"context"
	"fmt"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// CreateEditionCommand expresses the owner CreateEdition use case independently of transport.
type CreateEditionCommand struct {
	Currency string `json:"currency"`
}

// CreateEditionCommandHandler applies CreateEdition through one aggregate command transaction.
type CreateEditionCommandHandler struct {
	Editions a.AggregateCommandPort[d.EditionState]
}

func (h CreateEditionCommandHandler) Execute(ctx context.Context, m a.Metadata, c CreateEditionCommand) (a.Outcome, error) {
	return h.Editions.Execute(ctx, m, func(s a.Loaded[d.EditionState]) (a.Mutation[d.EditionState], error) {
		if s.Exists {
			return a.Mutation[d.EditionState]{}, core.Reject("already_exists", "The edition already exists")
		}
		edition, err := d.NewEdition(m.AggregateID, c.Currency)
		if err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		return a.Changed(edition.Snapshot(), "draft"), nil
	})
}

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
	Drinks   a.ProjectionPort[DrinkPublished]
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

// ChangePriceCommand expresses the owner ChangePrice use case independently of transport.
type ChangePriceCommand struct {
	Code  string `json:"code"`
	Minor int64  `json:"minor"`
}

// ChangePriceCommandHandler applies ChangePrice through one aggregate command transaction.
type ChangePriceCommandHandler struct {
	Editions a.AggregateCommandPort[d.EditionState]
}

func (h ChangePriceCommandHandler) Execute(ctx context.Context, m a.Metadata, c ChangePriceCommand) (a.Outcome, error) {
	return h.Editions.Execute(ctx, m, func(s a.Loaded[d.EditionState]) (a.Mutation[d.EditionState], error) {
		if !s.Exists {
			return a.Mutation[d.EditionState]{}, core.Reject("not_found", "The edition does not exist")
		}
		edition, err := d.RestoreEdition(s.State)
		if err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		if err = edition.ChangePrice(c.Code, c.Minor); err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		return a.Mutation[d.EditionState]{State: edition.Snapshot(), Changed: len(edition.Events()) > 0, Status: "draft"}, nil
	})
}

// PublishEditionCommandHandler applies PublishEdition through one aggregate command transaction.
type PublishEditionCommandHandler struct {
	Editions a.AggregateCommandPort[d.EditionState]
}

func (h PublishEditionCommandHandler) Execute(ctx context.Context, m a.Metadata, _ PublishEditionCommand) (a.Outcome, error) {
	return h.Editions.Execute(ctx, m, func(s a.Loaded[d.EditionState]) (a.Mutation[d.EditionState], error) {
		if !s.Exists {
			return a.Mutation[d.EditionState]{}, core.Reject("not_found", "The edition does not exist")
		}
		edition, err := d.RestoreEdition(s.State)
		if err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		if err = edition.Publish(); err != nil {
			return a.Mutation[d.EditionState]{}, err
		}
		payload := model.MenuPublished{EditionID: s.State.ID, Currency: s.State.Currency, Offers: []model.Offer{}}
		for _, offer := range edition.Snapshot().Offers {
			payload.Offers = append(payload.Offers, model.Offer{Code: offer.Code, DrinkID: offer.DrinkID, DrinkRevision: offer.DrinkRevision, Name: offer.Name, Minor: offer.Minor, Currency: offer.Currency})
		}
		return a.Changed(edition.Snapshot(), "published", a.Publication{Name: "menu.edition-published", Visibility: a.Public, Payload: payload}), nil
	})
}

// PublishEditionCommand expresses the owner use case independently of its transport.
type PublishEditionCommand struct{}
