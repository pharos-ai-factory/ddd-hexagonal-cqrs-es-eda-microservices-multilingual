package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ListDrinksQueryHandler preserves revision, absence and pagination policy.
type ListDrinksQueryHandler struct{ ReadRepository ports.DrinkReadRepository }

func (h ListDrinksQueryHandler) Execute(ctx context.Context, q model.ListDrinksQuery) (model.ListDrinksResponse, error) {
	var page a.Page[model.Drink]
	var err error
	if q.Page != nil {
		page, err = h.ReadRepository.Page(ctx, *q.Page)
	} else {
		page.Items, err = h.ReadRepository.List(ctx)
	}
	if err != nil {
		return model.ListDrinksResponse{}, err
	}
	result := model.Drinks{Items: make([]model.LoadedDrink, 0, len(page.Items)), Paged: q.Page != nil}
	for _, item := range page.Items {
		result.Items = append(result.Items, model.LoadedDrink{Exists: item.Exists, Version: item.Version, State: item.State})
	}
	if page.NextID != "" {
		result.NextID = &page.NextID
	}
	return model.ListDrinksResponse{Drinks: &result}, nil
}
