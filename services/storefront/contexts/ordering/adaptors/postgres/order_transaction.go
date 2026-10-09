package postgres

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/publications"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

// NewOrderTransaction coordinates the owner repository, receipts and outgoing intent.
func NewOrderTransaction(db *store.ContextDatabase) execution.Transaction[*d.Order] {
	return store.NewAggregateTransaction[d.OrderState, *d.Order](db, "order", NewOrderWriteRepository, publications.Order)
}
