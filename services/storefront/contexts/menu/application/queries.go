package application

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// DrinkQueries reads drink snapshots through the injected query port.
type DrinkQueries struct {
	Read a.PagedQueryPort[d.DrinkState]
}

func (h DrinkQueries) Get(ctx context.Context, id string) (a.Loaded[d.DrinkState], error) {
	return h.Read.Get(ctx, id)
}
func (h DrinkQueries) List(ctx context.Context) ([]a.Loaded[d.DrinkState], error) {
	return h.Read.List(ctx)
}

// EditionQueries reads edition snapshots through the injected query port.
type EditionQueries struct {
	Read a.PagedQueryPort[d.EditionState]
}

func (h EditionQueries) Get(ctx context.Context, id string) (a.Loaded[d.EditionState], error) {
	return h.Read.Get(ctx, id)
}
func (h EditionQueries) List(ctx context.Context) ([]a.Loaded[d.EditionState], error) {
	return h.Read.List(ctx)
}

func (h DrinkQueries) Page(ctx context.Context, request a.PageRequest) (a.Page[d.DrinkState], error) {
	return h.Read.Page(ctx, request)
}

func (h EditionQueries) Page(ctx context.Context, request a.PageRequest) (a.Page[d.EditionState], error) {
	return h.Read.Page(ctx, request)
}
