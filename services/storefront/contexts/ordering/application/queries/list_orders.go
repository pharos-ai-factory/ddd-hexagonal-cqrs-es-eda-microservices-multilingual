package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ListOrdersQuery optionally bounds the returned orders.
type ListOrdersQuery struct{ Page *a.PageRequest }

// ListOrdersQueryHandler reads orders with stable continuation semantics.
type ListOrdersQueryHandler struct{ Read ports.OrderReader }

func (h ListOrdersQueryHandler) Execute(ctx context.Context, q ListOrdersQuery) (a.Page[view.OrderView], error) {
	if q.Page != nil {
		return h.Read.Page(ctx, *q.Page)
	}
	items, err := h.Read.List(ctx)
	return a.Page[view.OrderView]{Items: items}, err
}
