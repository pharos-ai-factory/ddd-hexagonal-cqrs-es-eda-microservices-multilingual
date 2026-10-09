// Package publications maps owner domain facts to delivered messages.
package publications

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Edition maps only the facts this context has chosen to deliver.
func Edition(aggregate *d.MenuEdition) []a.Publication {
	var publications []a.Publication
	for _, fact := range aggregate.Events() {
		if fact.Name == "MenuEditionPublished" {
			s := fact.Data.(d.EditionState)
			payload := model.MenuPublished{EditionID: s.ID, Currency: s.Currency, Offers: []model.Offer{}}
			for _, offer := range s.Offers {
				payload.Offers = append(payload.Offers, model.Offer{Code: offer.Code, DrinkID: offer.DrinkID, DrinkRevision: offer.DrinkRevision, Name: offer.Name, Minor: offer.Minor, Currency: offer.Currency})
			}
			publications = append(publications, a.Publication{Name: "menu.edition-published", Visibility: a.Public, Payload: payload})
		}
	}
	return publications
}
