package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetEditionQuery identifies the edition requested by the caller.
type GetEditionQuery struct{ ID string }

// GetEditionQueryHandler reads one edition through its application capability.
type GetEditionQueryHandler struct{ Read ports.EditionReader }

func (h GetEditionQueryHandler) Execute(ctx context.Context, q GetEditionQuery) (a.Loaded[view.EditionView], error) {
	return h.Read.Get(ctx, q.ID)
}
