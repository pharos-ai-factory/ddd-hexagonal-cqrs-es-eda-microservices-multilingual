package application

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type DrinkQueries struct{ Read a.QueryPort[d.DrinkState] }

func (h DrinkQueries) Get(ctx context.Context, id string) (a.Loaded[d.DrinkState], error) {
	return h.Read.Get(ctx, id)
}
func (h DrinkQueries) List(ctx context.Context) ([]a.Loaded[d.DrinkState], error) {
	return h.Read.List(ctx)
}

type EditionQueries struct{ Read a.QueryPort[d.EditionState] }

func (h EditionQueries) Get(ctx context.Context, id string) (a.Loaded[d.EditionState], error) {
	return h.Read.Get(ctx, id)
}
func (h EditionQueries) List(ctx context.Context) ([]a.Loaded[d.EditionState], error) {
	return h.Read.List(ctx)
}
