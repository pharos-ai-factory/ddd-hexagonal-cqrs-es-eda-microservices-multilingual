package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/menu"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/shared"
)

// GetEditionQueryHandler preserves revision, absence and pagination policy.
type GetEditionQueryHandler struct{ ReadRepository ports.EditionReadRepository }

func (h GetEditionQueryHandler) Execute(ctx context.Context, q model.GetEditionQuery) (model.GetEditionResponse, error) {
	value, err := h.ReadRepository.Get(ctx, q.ID)
	if err != nil {
		return model.GetEditionResponse{}, err
	}
	if !value.Exists {
		return model.GetEditionResponse{Error: &shared.RequestError{Code: "not_found"}}, nil
	}
	return model.GetEditionResponse{Edition: &model.LoadedEdition{Exists: value.Exists, Version: value.Version, State: value.State}}, nil
}
