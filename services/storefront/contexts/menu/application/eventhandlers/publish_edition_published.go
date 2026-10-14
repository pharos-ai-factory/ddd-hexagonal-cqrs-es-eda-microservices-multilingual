// Package eventhandlers reacts to committed private facts.
package eventhandlers

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// MenuEditionPublishedDomainEventHandler publishes the integration contract through an atomic receipt/outbox port.
type MenuEditionPublishedDomainEventHandler struct {
	Publications a.DurablePublicationPort[model.MenuPublishedIntegrationEvent]
}

func (h MenuEditionPublishedDomainEventHandler) Handle(ctx context.Context, metadata a.Metadata, event d.MenuEditionPublishedDomainEvent) (a.Outcome, error) {
	s := event.State
	payload := model.MenuPublishedIntegrationEvent{EditionID: s.ID, Currency: s.Currency, Offers: []model.Offer{}}
	for _, offer := range s.Offers {
		payload.Offers = append(payload.Offers, model.Offer{Code: offer.Code, DrinkID: offer.DrinkID, DrinkRevision: offer.DrinkRevision, Name: offer.Name, Minor: offer.Minor, Currency: offer.Currency})
	}
	return h.Publications.Publish(ctx, metadata, payload)
}
