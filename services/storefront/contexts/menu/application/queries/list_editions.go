package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ListEditionsQuery optionally bounds the returned editions.
type ListEditionsQuery struct{ Page *a.PageRequest }

// ListEditionsQueryHandler reads editions with stable continuation semantics.
type ListEditionsQueryHandler struct{ ReadRepository ports.EditionReadRepository }

func (h ListEditionsQueryHandler) Execute(ctx context.Context, q ListEditionsQuery) (a.Page[view.EditionView], error) {
	if q.Page != nil {
		return h.ReadRepository.Page(ctx, *q.Page)
	}
	items, err := h.ReadRepository.List(ctx)
	return a.Page[view.EditionView]{Items: items}, err
}
