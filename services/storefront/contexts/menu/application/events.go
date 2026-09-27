package application

import (
	"context"
	"fmt"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type ProjectDrinkHandler struct {
	Directory a.ProjectionPort[model.DrinkPublished]
}

func (h ProjectDrinkHandler) Handle(ctx context.Context, m a.Metadata, event model.DrinkPublished) (a.Outcome, error) {
	key := fmt.Sprintf("%s/%d", event.DrinkID, event.Revision)
	return a.Outcome{}, h.Directory.Record(ctx, m, key, event.Revision, event)
}
