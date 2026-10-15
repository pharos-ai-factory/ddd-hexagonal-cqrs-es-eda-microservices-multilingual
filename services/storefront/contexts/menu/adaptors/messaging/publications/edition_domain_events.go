// Package publications stages owner-private domain events.
package publications

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/names"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Edition selects committed facts for independent private subscriptions.
func Edition(aggregate *d.MenuEdition) []a.Publication {
	var result []a.Publication
	for _, fact := range aggregate.Events() {
		if event, ok := fact.(d.MenuEditionPublishedDomainEvent); ok {
			result = append(result, a.Publication{Name: names.MenuDomainEventEditionPublished.String(), Visibility: a.Private, Payload: event})
		}
	}
	return result
}
