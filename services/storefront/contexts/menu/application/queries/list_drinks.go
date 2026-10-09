package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ListDrinksQuery optionally bounds the returned drinks.
type ListDrinksQuery struct{ Page *a.PageRequest }

// ListDrinksQueryHandler reads drinks with stable continuation semantics.
type ListDrinksQueryHandler struct{ Read ports.DrinkReader }

func (h ListDrinksQueryHandler) Execute(ctx context.Context, q ListDrinksQuery) (a.Page[view.DrinkView], error) {
	if q.Page != nil {
		return h.Read.Page(ctx, *q.Page)
	}
	items, err := h.Read.List(ctx)
	return a.Page[view.DrinkView]{Items: items}, err
}
