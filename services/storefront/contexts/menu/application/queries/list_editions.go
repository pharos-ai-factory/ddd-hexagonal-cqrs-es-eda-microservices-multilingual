package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ListEditionsQueryHandler preserves revision, absence and pagination policy.
type ListEditionsQueryHandler struct{ ReadRepository ports.EditionReadRepository }

func (h ListEditionsQueryHandler) Execute(ctx context.Context, q model.ListEditionsQuery) (model.ListEditionsResponse, error) {
	var page a.Page[model.Edition]
	var err error
	if q.Page != nil {
		page, err = h.ReadRepository.Page(ctx, *q.Page)
	} else {
		page.Items, err = h.ReadRepository.List(ctx)
	}
	if err != nil {
		return model.ListEditionsResponse{}, err
	}
	result := model.Editions{Items: make([]model.LoadedEdition, 0, len(page.Items)), Paged: q.Page != nil}
	for _, item := range page.Items {
		result.Items = append(result.Items, model.LoadedEdition{Exists: item.Exists, Version: item.Version, State: item.State})
	}
	if page.NextID != "" {
		result.NextID = &page.NextID
	}
	return model.ListEditionsResponse{Editions: &result}, nil
}
