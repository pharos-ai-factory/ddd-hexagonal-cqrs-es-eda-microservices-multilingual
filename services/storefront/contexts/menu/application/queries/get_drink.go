package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetDrinkQuery identifies the drink requested by the caller.
type GetDrinkQuery struct{ ID string }

// GetDrinkQueryHandler reads one drink through its application capability.
type GetDrinkQueryHandler struct{ Read ports.DrinkReader }

func (h GetDrinkQueryHandler) Execute(ctx context.Context, q GetDrinkQuery) (a.Loaded[view.DrinkView], error) {
	return h.Read.Get(ctx, q.ID)
}
