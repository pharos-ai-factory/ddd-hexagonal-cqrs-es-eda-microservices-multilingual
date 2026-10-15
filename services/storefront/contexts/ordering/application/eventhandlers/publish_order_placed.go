// Package eventhandlers reacts to committed private facts.
package eventhandlers

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderPlacedDomainEventHandler publishes the integration contract through an atomic receipt/outbox port.
type OrderPlacedDomainEventHandler struct {
	Publications a.IntegrationEventBusPort[model.OrderPlacedIntegrationEvent]
}

func (h OrderPlacedDomainEventHandler) Handle(ctx context.Context, metadata a.Metadata, event d.OrderPlacedDomainEvent) (a.Outcome, error) {
	s := event.State
	payload := model.OrderPlacedIntegrationEvent{OrderID: s.ID, CustomerID: s.CustomerID, EditionID: s.EditionID, Currency: s.Currency, Lines: []model.Line{}}
	for _, line := range s.Lines {
		payload.Lines = append(payload.Lines, model.Line{ID: line.ID, OfferCode: line.Selection.OfferCode, Name: line.Selection.Name, Quantity: line.Quantity, Minor: line.Selection.Minor})
	}
	return h.Publications.Publish(ctx, metadata, payload)
}
