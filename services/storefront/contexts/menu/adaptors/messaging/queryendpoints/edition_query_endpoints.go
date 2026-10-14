package queries

import (
	"context"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// EditionQueryEndpoints maps transport arguments into named application queries.
type EditionQueryEndpoints struct {
	GetHandler  q.GetEditionQueryHandler
	ListHandler q.ListEditionsQueryHandler
}

func (h EditionQueryEndpoints) Get(ctx context.Context, id string) (a.Loaded[view.EditionView], error) {
	return h.GetHandler.Execute(ctx, q.GetEditionQuery{ID: id})
}
func (h EditionQueryEndpoints) List(ctx context.Context) ([]a.Loaded[view.EditionView], error) {
	result, err := h.ListHandler.Execute(ctx, q.ListEditionsQuery{})
	return result.Items, err
}
func (h EditionQueryEndpoints) Page(ctx context.Context, page a.PageRequest) (a.Page[view.EditionView], error) {
	return h.ListHandler.Execute(ctx, q.ListEditionsQuery{Page: &page})
}
