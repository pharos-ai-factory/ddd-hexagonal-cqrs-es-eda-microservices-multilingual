package postgres

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

// NewEditionReadRepository constructs the PostgreSQL repository for validated application views.
func NewEditionReadRepository(db *store.ContextDatabase) ports.EditionReadRepository {
	return store.MappedReadRepository[d.EditionState, view.EditionView]{Source: store.NewSnapshotReadRepository[d.EditionState](db, "edition"), View: editionView}
}
func editionView(s d.EditionState) (view.EditionView, error) {
	root, err := d.RestoreEdition(s)
	if err != nil {
		return view.EditionView{}, err
	}
	s = root.Snapshot()
	offers := make([]view.OfferView, 0, len(s.Offers))
	for _, o := range s.Offers {
		offers = append(offers, view.OfferView{Code: o.Code, DrinkID: o.DrinkID, DrinkRevision: o.DrinkRevision, Name: o.Name, Minor: o.Minor, Currency: o.Currency})
	}
	return view.EditionView{ID: s.ID, Currency: s.Currency, Status: s.Status, Offers: offers}, nil
}
