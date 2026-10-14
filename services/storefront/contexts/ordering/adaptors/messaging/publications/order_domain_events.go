// Package publications stages owner-private domain events.
package publications

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/names"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Order selects committed facts for independent private subscriptions.
func Order(aggregate *d.Order) []a.Publication {
	var result []a.Publication
	for _, fact := range aggregate.Events() {
		if event, ok := fact.(d.OrderPlacedDomainEvent); ok {
			result = append(result, a.Publication{Name: names.OrderingDomainEventOrderPlaced.String(), Visibility: a.Private, Payload: event})
		}
	}
	return result
}
