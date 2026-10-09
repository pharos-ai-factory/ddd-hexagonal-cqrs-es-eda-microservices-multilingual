// Package publications maps owner domain facts to delivered messages.
package publications

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Order maps only the facts this context has chosen to deliver.
func Order(aggregate *d.Order) []a.Publication {
	var publications []a.Publication
	for _, fact := range aggregate.Events() {
		if fact.Name == "OrderPlaced" {
			s := fact.Data.(d.OrderState)
			payload := model.OrderPlaced{OrderID: s.ID, CustomerID: s.CustomerID, EditionID: s.EditionID, Currency: s.Currency, Lines: []model.Line{}}
			for _, line := range s.Lines {
				payload.Lines = append(payload.Lines, model.Line{ID: line.ID, OfferCode: line.Selection.OfferCode, Name: line.Selection.Name, Quantity: line.Quantity, Minor: line.Selection.Minor})
			}
			publications = append(publications, a.Publication{Name: "ordering.order-placed", Visibility: a.Public, Payload: payload})
		}
	}
	return publications
}
