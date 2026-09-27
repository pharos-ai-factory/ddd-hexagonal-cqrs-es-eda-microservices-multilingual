package application

import (
	"context"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type ProjectMenuHandler struct {
	Directory a.ProjectionPort[model.MenuPublished]
}

func (h ProjectMenuHandler) Handle(ctx context.Context, m a.Metadata, event model.MenuPublished) (a.Outcome, error) {
	// Each edition is published once; later menus have a new edition identity.
	return a.Outcome{}, h.Directory.Record(ctx, m, event.EditionID, 1, event)
}
