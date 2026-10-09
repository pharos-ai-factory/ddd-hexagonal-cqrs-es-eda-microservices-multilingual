package postgres

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/publications"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

// NewEditionTransaction coordinates the owner repository, receipts and outgoing intent.
func NewEditionTransaction(db *store.ContextDatabase) execution.Transaction[*d.MenuEdition] {
	return store.NewAggregateTransaction[d.EditionState, *d.MenuEdition](db, "edition", NewEditionWriteRepository, publications.Edition)
}
