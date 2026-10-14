// Package publications maps owner domain facts to delivered messages.
package publications

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/names"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Drink maps only the facts this context has chosen to deliver.
func Drink(aggregate *d.Drink) []a.Publication {
	var publications []a.Publication
	for _, fact := range aggregate.Events() {
		if event, ok := fact.(d.DrinkPublishedDomainEvent); ok {
			s := event.State
			publications = append(publications, a.Publication{Name: names.MenuDomainEventDrinkPublished.String(), Visibility: a.Private, Payload: app.DrinkPublishedDomainEvent{DrinkID: s.ID, Name: s.Name, Revision: s.Revision}})
		}
	}
	return publications
}
