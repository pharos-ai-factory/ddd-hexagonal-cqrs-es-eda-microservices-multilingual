package eventhandlers

import (
	"context"
	"testing"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type realtimeOrderReader struct {
	ports.OrderReadRepository
	loaded a.Loaded[view.OrderView]
}

func (r *realtimeOrderReader) Get(context.Context, string) (a.Loaded[view.OrderView], error) {
	return r.loaded, nil
}

type realtimeRecorder struct{ events []a.RealtimeEvent }

func (b *realtimeRecorder) Publish(_ context.Context, event a.RealtimeEvent) error {
	b.events = append(b.events, event)
	return nil
}

func TestRealtimeOrderRefreshUsesReadRevisionAndRequiresCommittedView(t *testing.T) {
	id := "00000000-0000-4000-8000-000000000001"
	reader := &realtimeOrderReader{loaded: a.Loaded[view.OrderView]{
		Exists: true, Version: 4, State: view.OrderView{ID: id, Status: "placed"},
	}}
	bus := &realtimeRecorder{}
	handler := OrderPlacedRealtimeDomainEventHandler{Orders: reader, Realtime: bus}
	source := a.Metadata{AggregateID: id, SourceVersion: 2}
	event := d.OrderPlacedDomainEvent{State: d.OrderState{ID: id}}
	outcome, err := handler.Handle(t.Context(), source, event)
	if err != nil || outcome.Version != 4 || len(bus.events) != 1 {
		t.Fatalf("refresh: %+v %v", outcome, err)
	}
	if bus.events[0].Revision != 4 || bus.events[0].Source.SourceVersion != 2 {
		t.Fatal("source revision replaced the read revision")
	}
	reader.loaded.Exists = false
	if _, err = handler.Handle(t.Context(), source, event); err == nil || len(bus.events) != 1 {
		t.Fatal("missing committed view should fail without publication")
	}
}
