package projections

import (
	"context"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// MenuDirectoryProjectionHandler updates the consumer-owned projection without aggregate mutation.
type MenuDirectoryProjectionHandler struct {
	Directory a.ProjectionPort[model.MenuPublishedIntegrationEvent]
}

func (h MenuDirectoryProjectionHandler) Handle(ctx context.Context, m a.Metadata, event model.MenuPublishedIntegrationEvent) (a.Outcome, error) {
	// Each edition is published once; later menus have a new edition identity.
	return a.Outcome{}, h.Directory.Record(ctx, m, event.EditionID, 1, event)
}
