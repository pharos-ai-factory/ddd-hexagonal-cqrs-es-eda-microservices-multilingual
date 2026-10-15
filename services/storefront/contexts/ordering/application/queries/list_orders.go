package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ListOrdersQueryHandler preserves revision, absence and pagination policy.
type ListOrdersQueryHandler struct{ ReadRepository ports.OrderReadRepository }

func (h ListOrdersQueryHandler) Execute(ctx context.Context, q model.ListOrdersQuery) (model.ListOrdersResponse, error) {
	var page a.Page[model.Order]
	var err error
	if q.Page != nil {
		page, err = h.ReadRepository.Page(ctx, *q.Page)
	} else {
		page.Items, err = h.ReadRepository.List(ctx)
	}
	if err != nil {
		return model.ListOrdersResponse{}, err
	}
	result := model.Orders{Items: make([]model.LoadedOrder, 0, len(page.Items)), Paged: q.Page != nil}
	for _, item := range page.Items {
		result.Items = append(result.Items, model.LoadedOrder{Exists: item.Exists, Version: item.Version, State: item.State})
	}
	if page.NextID != "" {
		result.NextID = &page.NextID
	}
	return model.ListOrdersResponse{Orders: &result}, nil
}
