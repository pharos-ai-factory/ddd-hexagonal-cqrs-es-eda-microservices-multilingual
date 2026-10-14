//go:build integration

package postgres

import (
	"testing"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

func TestProposedRootMustIdentifyTheCommandTarget(t *testing.T) {
	type root struct {
		ID string `json:"id"`
	}
	db, owner := fixtureDB(t), admin(t)
	id := NewID()
	clean(t, owner, id)
	m := metadata(id, 0)
	store := snapshotDecisions[root](db, "test_counter")
	if _, err := store.Execute(t.Context(), m, func(a.Loaded[root]) (a.Mutation[root], error) {
		return a.Changed(root{ID: NewID()}, "active"), nil
	}); err == nil {
		t.Fatal("persisted a proposal identifying another root")
	}
	if loaded, err := NewSnapshotReadRepository[root](db, "test_counter").Get(t.Context(), id); err != nil || loaded.Exists {
		t.Fatalf("failed proposal changed stored authority: %+v %v", loaded, err)
	}
	for _, table := range []string{"command_receipts", "outbox_events", "realtime_publications"} {
		var count int
		if err := owner.QueryRow(t.Context(), "SELECT count(*) FROM cafe."+table+" WHERE aggregate_id=$1", id).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatal("failed proposal left durable intent in " + table)
		}
	}
	outcome, err := store.Execute(t.Context(), m, func(a.Loaded[root]) (a.Mutation[root], error) {
		return a.Changed(root{ID: id}, "active"), nil
	})
	if err != nil || outcome.Version != 1 || outcome.Rejection != nil {
		t.Fatalf("original attempt did not recover: %+v %v", outcome, err)
	}
}
