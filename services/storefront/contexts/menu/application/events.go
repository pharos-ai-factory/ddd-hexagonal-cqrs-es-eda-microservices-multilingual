package application

import (
	"context"
	"fmt"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// DrinkDirectoryProjectionHandler updates the consumer-owned projection without aggregate mutation.
type DrinkDirectoryProjectionHandler struct {
	Directory a.ProjectionPort[DrinkPublished]
}

func (h DrinkDirectoryProjectionHandler) Handle(ctx context.Context, m a.Metadata, event DrinkPublished) (a.Outcome, error) {
	key := fmt.Sprintf("%s/%d", event.DrinkID, event.Revision)
	return a.Outcome{}, h.Directory.Record(ctx, m, key, event.Revision, event)
}

// DrinkPublished is an owner-private application delivery value.
type DrinkPublished struct {
	DrinkID  string `json:"drinkId"`
	Name     string `json:"name"`
	Revision uint64 `json:"revision"`
}
