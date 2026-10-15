package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/shared"
)

// GetOrderQueryHandler preserves revision, absence and pagination policy.
type GetOrderQueryHandler struct{ ReadRepository ports.OrderReadRepository }

func (h GetOrderQueryHandler) Execute(ctx context.Context, q model.GetOrderQuery) (model.GetOrderResponse, error) {
	value, err := h.ReadRepository.Get(ctx, q.ID)
	if err != nil {
		return model.GetOrderResponse{}, err
	}
	if !value.Exists {
		return model.GetOrderResponse{Error: &shared.RequestError{Code: "not_found"}}, nil
	}
	return model.GetOrderResponse{Order: &model.LoadedOrder{Exists: value.Exists, Version: value.Version, State: value.State}}, nil
}
