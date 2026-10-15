package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/shared"
)

// GetDrinkQueryHandler preserves revision, absence and pagination policy.
type GetDrinkQueryHandler struct{ ReadRepository ports.DrinkReadRepository }

func (h GetDrinkQueryHandler) Execute(ctx context.Context, q model.GetDrinkQuery) (model.GetDrinkResponse, error) {
	value, err := h.ReadRepository.Get(ctx, q.ID)
	if err != nil {
		return model.GetDrinkResponse{}, err
	}
	if !value.Exists {
		return model.GetDrinkResponse{Error: &shared.RequestError{Code: "not_found"}}, nil
	}
	return model.GetDrinkResponse{Drink: &model.LoadedDrink{Exists: value.Exists, Version: value.Version, State: value.State}}, nil
}
