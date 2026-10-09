package main

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	"go.uber.org/fx"
	"testing"
)

func TestCompleteFxGraphWithoutInfrastructure(t *testing.T) {
	if err := fx.ValidateApp(composition(), fx.Supply(&support.ResourceScope{}), fx.NopLogger); err != nil {
		t.Fatal(err)
	}
}
