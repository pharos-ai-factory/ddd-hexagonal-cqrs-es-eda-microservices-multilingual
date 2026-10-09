package postgres

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

// NewDrinkWriteRepository maps the scoped PostgreSQL snapshot repository to owner aggregates.
func NewDrinkWriteRepository(source a.WriteRepository[d.DrinkState]) ports.DrinkWriteRepository {
	return store.MappedWriteRepository[d.DrinkState, *d.Drink]{Source: source, Restore: d.RestoreDrink, Snapshot: (*d.Drink).Snapshot}
}
