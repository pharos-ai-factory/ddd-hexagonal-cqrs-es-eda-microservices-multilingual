package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ListOrdersQueryHandler reads orders with stable continuation semantics.
type ListOrdersQueryHandler struct{ ReadRepository ports.OrderReadRepository }

func (h ListOrdersQueryHandler) Execute(ctx context.Context, q orderingcontracts.ListOrdersQuery) (a.Page[view.OrderView], error) {
	if q.Page != nil {
		return h.ReadRepository.Page(ctx, *q.Page)
	}
	items, err := h.ReadRepository.List(ctx)
	return a.Page[view.OrderView]{Items: items}, err
}
