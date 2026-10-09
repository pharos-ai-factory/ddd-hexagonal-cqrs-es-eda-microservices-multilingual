package commands

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// PublishEditionCommand expresses the owner use case independently of its transport.
type PublishEditionCommand struct{}

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
