package postgres

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

func OrderCommands(db *store.ContextDatabase) *store.AggregateCommandStore[d.OrderState] {
	return store.Command[d.OrderState](db, "order")
}
