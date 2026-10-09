package postgres

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

// NewDrinkReadRepository constructs the PostgreSQL repository for validated application views.
func NewDrinkReadRepository(db *store.ContextDatabase) ports.DrinkReadRepository {
	return store.MappedReadRepository[d.DrinkState, view.DrinkView]{Source: store.NewSnapshotReadRepository[d.DrinkState](db, "drink"), View: drinkView}
}
func drinkView(s d.DrinkState) (view.DrinkView, error) {
	root, err := d.RestoreDrink(s)
	if err != nil {
		return view.DrinkView{}, err
	}
	s = root.Snapshot()
	return view.DrinkView{ID: s.ID, Name: s.Name, Revision: s.Revision, Published: s.Published}, nil
}
