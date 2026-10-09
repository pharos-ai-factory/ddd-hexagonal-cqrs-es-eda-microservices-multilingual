package postgres

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/publications"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

// NewDrinkTransaction coordinates the owner repository, receipts and outgoing intent.
func NewDrinkTransaction(db *store.ContextDatabase) execution.Transaction[*d.Drink] {
	return store.NewAggregateTransaction[d.DrinkState, *d.Drink](db, "drink", NewDrinkWriteRepository, publications.Drink)
}
