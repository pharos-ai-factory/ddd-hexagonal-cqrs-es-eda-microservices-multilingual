package messaging

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"testing"
)

func TestMenuEvidenceIsCheckedBeforeTheProjectionRuns(t *testing.T) {
	for _, offers := range [][]model.Offer{
		{{Code: "TEA", Currency: "EUR"}, {Code: "TEA", Currency: "EUR"}},
		{{Code: "TEA", Currency: "GBP"}},
	} {
		called := false
		receive := MenuPublished(func(context.Context, a.Metadata, model.MenuPublished) (a.Outcome, error) {
			called = true
			return a.Outcome{}, nil
		})
		if _, err := receive(context.Background(), a.Metadata{}, model.MenuPublished{Currency: "EUR", Offers: offers}); err == nil || called {
			t.Fatal("inconsistent evidence reached the projection")
		}
	}
}
