package queries

import (
	"context"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// DrinkQueryEndpoints maps transport arguments into named application queries.
type DrinkQueryEndpoints struct {
	GetHandler  q.GetDrinkQueryHandler
	ListHandler q.ListDrinksQueryHandler
}

func (h DrinkQueryEndpoints) Get(ctx context.Context, id string) (a.Loaded[view.DrinkView], error) {
	return h.GetHandler.Execute(ctx, q.GetDrinkQuery{ID: id})
}
func (h DrinkQueryEndpoints) List(ctx context.Context) ([]a.Loaded[view.DrinkView], error) {
	result, err := h.ListHandler.Execute(ctx, q.ListDrinksQuery{})
	return result.Items, err
}
func (h DrinkQueryEndpoints) Page(ctx context.Context, page a.PageRequest) (a.Page[view.DrinkView], error) {
	return h.ListHandler.Execute(ctx, q.ListDrinksQuery{Page: &page})
}
