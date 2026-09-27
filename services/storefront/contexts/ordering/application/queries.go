package application

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type OrderingQueries struct{ Read a.QueryPort[d.State] }

func (h OrderingQueries) Get(ctx context.Context, id string) (a.Loaded[d.State], error) {
	return h.Read.Get(ctx, id)
}
func (h OrderingQueries) List(ctx context.Context) ([]a.Loaded[d.State], error) {
	return h.Read.List(ctx)
}
