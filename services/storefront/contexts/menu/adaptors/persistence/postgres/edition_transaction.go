package postgres

import (
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/publications"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
)

// NewEditionTransaction coordinates the owner repository, receipts and outgoing intent.
func NewEditionTransaction(db *store.ContextDatabase) execution.Transaction[*d.MenuEdition] {
	return store.NewAggregateTransaction[d.EditionState, *d.MenuEdition](db, "edition", NewEditionWriteRepository, publications.Edition)
}
