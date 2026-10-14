package command

import (
	"context"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"testing"
)

type repositoryProbe struct{ saved bool }

func (r *repositoryProbe) Get(context.Context, string) (a.Loaded[int], error) {
	return a.Loaded[int]{}, nil
}
func (r *repositoryProbe) Save(context.Context, int) error { r.saved = true; return nil }
func TestAggregateSaveRejectsIntegrationPublication(t *testing.T) {
	source := &repositoryProbe{}
	repository := EventRecordingRepository[int]{Source: source, Map: func(int) []a.Publication {
		return []a.Publication{{Name: "ordering.order-placed", Visibility: a.Public}}
	}}
	if err := repository.Save(t.Context(), 1); err == nil || source.saved {
		t.Fatal("aggregate bypassed its domain subscription")
	}
	repository.Map = func(int) []a.Publication {
		return []a.Publication{{Name: "ordering.order-placed.domain", Visibility: a.Private}}
	}
	if err := repository.Save(t.Context(), 1); err != nil || !source.saved || len(repository.Publications) != 1 {
		t.Fatal("private event was not staged")
	}
}
