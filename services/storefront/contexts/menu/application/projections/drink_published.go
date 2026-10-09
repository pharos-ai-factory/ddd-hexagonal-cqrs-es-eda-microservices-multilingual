package projections

import (
	"context"
	"fmt"
	events "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// DrinkDirectoryProjectionHandler updates the consumer-owned projection without aggregate mutation.
type DrinkDirectoryProjectionHandler struct {
	Directory a.ProjectionPort[events.DrinkPublished]
}

func (h DrinkDirectoryProjectionHandler) Handle(ctx context.Context, m a.Metadata, event events.DrinkPublished) (a.Outcome, error) {
	key := fmt.Sprintf("%s/%d", event.DrinkID, event.Revision)
	return a.Outcome{}, h.Directory.Record(ctx, m, key, event.Revision, event)
}
