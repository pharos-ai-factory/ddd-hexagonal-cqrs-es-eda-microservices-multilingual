package postgres

import (
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// NewOrderWriteRepository maps the scoped PostgreSQL snapshot repository to owner aggregates.
func NewOrderWriteRepository(source a.WriteRepository[d.OrderState]) ports.OrderWriteRepository {
	return store.MappedWriteRepository[d.OrderState, *d.Order]{Source: source, Restore: d.Restore, Snapshot: (*d.Order).Snapshot}
}
