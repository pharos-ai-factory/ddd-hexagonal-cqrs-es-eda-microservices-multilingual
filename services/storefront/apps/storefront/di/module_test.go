package di_test

import (
	"testing"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/storefront/di"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	"go.uber.org/fx"
)

func TestCompleteFxGraphWithoutInfrastructure(t *testing.T) {
	if err := fx.ValidateApp(di.Module(), fx.Supply(&support.ResourceScope{}), fx.NopLogger); err != nil {
		t.Fatal(err)
	}
}
