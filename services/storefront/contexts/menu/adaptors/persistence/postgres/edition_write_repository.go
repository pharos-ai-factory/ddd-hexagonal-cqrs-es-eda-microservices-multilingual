package postgres

import (
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// NewEditionWriteRepository maps the scoped PostgreSQL snapshot repository to owner aggregates.
func NewEditionWriteRepository(source a.WriteRepository[d.EditionState]) ports.EditionWriteRepository {
	return store.MappedWriteRepository[d.EditionState, *d.MenuEdition]{Source: source, Restore: d.RestoreEdition, Snapshot: (*d.MenuEdition).Snapshot}
}
