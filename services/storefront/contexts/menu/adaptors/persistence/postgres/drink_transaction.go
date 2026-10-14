package postgres

import (
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/publications"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
)

// NewDrinkTransaction coordinates the owner repository, receipts and outgoing intent.
func NewDrinkTransaction(db *store.ContextDatabase) execution.Transaction[*d.Drink] {
	return store.NewAggregateTransaction[d.DrinkState, *d.Drink](db, "drink", NewDrinkWriteRepository, publications.Drink)
}
