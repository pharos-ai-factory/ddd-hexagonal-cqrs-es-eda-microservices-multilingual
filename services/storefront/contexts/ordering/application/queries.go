package application

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderingQueries reads order snapshots through the injected query port.
type OrderingQueries struct {
	Read a.PagedQueryPort[d.OrderState]
}

func (h OrderingQueries) Get(ctx context.Context, id string) (a.Loaded[d.OrderState], error) {
	return h.Read.Get(ctx, id)
}
func (h OrderingQueries) List(ctx context.Context) ([]a.Loaded[d.OrderState], error) {
	return h.Read.List(ctx)
}

func (h OrderingQueries) Page(ctx context.Context, request a.PageRequest) (a.Page[d.OrderState], error) {
	return h.Read.Page(ctx, request)
}
