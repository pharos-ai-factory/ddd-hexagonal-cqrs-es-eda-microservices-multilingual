package postgres

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

func DrinkCommands(db *store.ContextDatabase) *store.AggregateCommandStore[d.DrinkState] {
	return store.Command[d.DrinkState](db, "drink")
}
func DrinkQueries(db *store.ContextDatabase) *store.AggregateQueries[d.DrinkState] {
	return store.Query[d.DrinkState](db, "drink")
}
func EditionCommands(db *store.ContextDatabase) *store.AggregateCommandStore[d.EditionState] {
	return store.Command[d.EditionState](db, "edition")
}
func EditionQueries(db *store.ContextDatabase) *store.AggregateQueries[d.EditionState] {
	return store.Query[d.EditionState](db, "edition")
}
