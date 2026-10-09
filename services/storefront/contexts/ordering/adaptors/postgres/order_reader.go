package postgres

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

// OrderReader restores stored authority before selecting the application view.
func OrderReader(db *store.ContextDatabase) ports.OrderReader {
	return store.MappedQueries[d.OrderState, view.OrderView]{Source: store.Query[d.OrderState](db, "order"), View: orderView}
}
func orderView(s d.OrderState) (view.OrderView, error) {
	root, err := d.Restore(s)
	if err != nil {
		return view.OrderView{}, err
	}
	s = root.Snapshot()
	lines := make([]view.LineView, 0, len(s.Lines))
	for _, l := range s.Lines {
		lines = append(lines, view.LineView{ID: l.ID, Quantity: l.Quantity, Selection: view.SelectionView{OfferCode: l.Selection.OfferCode, Name: l.Selection.Name, Minor: l.Selection.Minor}})
	}
	return view.OrderView{ID: s.ID, CustomerID: s.CustomerID, EditionID: s.EditionID, Currency: s.Currency, Status: s.Status, Lines: lines}, nil
}
