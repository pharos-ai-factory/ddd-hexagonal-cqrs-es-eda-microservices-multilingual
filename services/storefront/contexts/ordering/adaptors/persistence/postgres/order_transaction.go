package postgres

import (
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging/publications"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
)

// NewOrderTransaction coordinates the owner repository, receipts and outgoing intent.
func NewOrderTransaction(db *store.ContextDatabase) execution.Transaction[*d.Order] {
	return store.NewAggregateTransaction[d.OrderState, *d.Order](db, "order", NewOrderWriteRepository, publications.Order)
}
