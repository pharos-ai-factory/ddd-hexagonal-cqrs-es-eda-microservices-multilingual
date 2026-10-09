package main

import (
	"go.uber.org/fx"
	"testing"
)

func TestCompleteFxGraphWithoutInfrastructure(t *testing.T) {
	if err := fx.ValidateApp(composition(), fx.NopLogger); err != nil {
		t.Fatal(err)
	}
}
