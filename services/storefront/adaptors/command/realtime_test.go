package command

import (
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"testing"
)

func TestRealtimeBusCapturesExplicitValuesAndExpires(t *testing.T) {
	bus := &RealtimeBuffer{active: true}
	state := map[string]any{"status": "draft"}
	message := a.SnapshotMessage(a.ContextTarget("ordering"), state)
	if err := bus.Publish(t.Context(), message); err != nil {
		t.Fatal(err)
	}
	state["status"] = "placed"
	if bus.Events[0].Snapshot.(map[string]any)["status"] != "draft" {
		t.Fatal("publication retained mutable state")
	}
	if err := bus.Publish(t.Context(), a.RealtimeMessage{}); err == nil {
		t.Fatal("empty publication accepted")
	}
	bus.Close()
	if err := bus.Publish(t.Context(), a.SnapshotMessage(a.ContextTarget("ordering"), state)); err == nil {
		t.Fatal("escaped bus stayed active")
	}
}
