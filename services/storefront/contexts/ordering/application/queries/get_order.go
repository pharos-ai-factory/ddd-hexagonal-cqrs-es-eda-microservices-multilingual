package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetOrderQueryHandler reads one order through its application capability.
type GetOrderQueryHandler struct{ ReadRepository ports.OrderReadRepository }

func (h GetOrderQueryHandler) Execute(ctx context.Context, q orderingcontracts.GetOrderQuery) (a.Loaded[view.OrderView], error) {
	return h.ReadRepository.Get(ctx, q.ID)
}
