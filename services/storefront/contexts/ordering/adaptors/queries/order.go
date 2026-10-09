package queries

import (
	"context"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/queries"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderQueryEndpoints maps transport arguments into named application queries.
type OrderQueryEndpoints struct {
	GetHandler  q.GetOrderQueryHandler
	ListHandler q.ListOrdersQueryHandler
}

func (h OrderQueryEndpoints) Get(ctx context.Context, id string) (a.Loaded[view.OrderView], error) {
	return h.GetHandler.Execute(ctx, q.GetOrderQuery{ID: id})
}
func (h OrderQueryEndpoints) List(ctx context.Context) ([]a.Loaded[view.OrderView], error) {
	result, err := h.ListHandler.Execute(ctx, q.ListOrdersQuery{})
	return result.Items, err
}
func (h OrderQueryEndpoints) Page(ctx context.Context, page a.PageRequest) (a.Page[view.OrderView], error) {
	return h.ListHandler.Execute(ctx, q.ListOrdersQuery{Page: &page})
}
