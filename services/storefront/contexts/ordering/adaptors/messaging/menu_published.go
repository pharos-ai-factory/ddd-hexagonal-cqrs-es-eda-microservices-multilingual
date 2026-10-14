package messaging

import (
	"context"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// MenuPublished checks immutable menu evidence before updating Ordering's projection.
func MenuPublished(handle func(context.Context, a.Metadata, model.MenuPublished) (a.Outcome, error)) func(context.Context, a.Metadata, model.MenuPublished) (a.Outcome, error) {
	return func(ctx context.Context, metadata a.Metadata, event model.MenuPublished) (a.Outcome, error) {
		seen := map[string]bool{}
		for _, offer := range event.Offers {
			if seen[offer.Code] || offer.Currency != event.Currency {
				return a.Outcome{}, fmt.Errorf("inconsistent published menu evidence")
			}
			seen[offer.Code] = true
		}
		return handle(ctx, metadata, event)
	}
}
