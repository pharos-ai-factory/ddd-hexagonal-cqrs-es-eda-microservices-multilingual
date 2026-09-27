package postgres

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

func OrderCommands(db *store.Database) *store.CommandStore[d.State] {
	return store.Command[d.State](db, "order")
}
func OrderQueries(db *store.Database) *store.Queries[d.State] {
	return store.Query[d.State](db, "order")
}
