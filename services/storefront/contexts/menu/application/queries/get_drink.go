package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetDrinkQueryHandler reads one drink through its application capability.
type GetDrinkQueryHandler struct{ ReadRepository ports.DrinkReadRepository }

func (h GetDrinkQueryHandler) Execute(ctx context.Context, q menucontracts.GetDrinkQuery) (a.Loaded[view.DrinkView], error) {
	return h.ReadRepository.Get(ctx, q.ID)
}
