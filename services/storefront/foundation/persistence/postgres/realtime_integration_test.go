//go:build integration

package postgres

import (
	"errors"
	"testing"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

func TestRealtimeIntentIsAtomicAndDispatchIsFenced(t *testing.T) {
	db, owner := fixtureDB(t), admin(t)
	id := NewID()
	clean(t, owner, id)
	encoder := func(_, _, _, _ string, _ uint64, state []byte) ([]byte, error) {
		return state, nil
	}
	if err := db.EnableRealtime(t.Context(), encoder); err != nil {
		t.Fatal(err)
	}
	db.realtime = func(_, _, _, _ string, _ uint64, _ []byte) ([]byte, error) {
		return nil, errors.New("encoding failed after state write")
	}
	m := metadata(id, 0)
	if _, err := Command[counter](db, "test_counter").Execute(t.Context(), m, increment); err == nil {
		t.Fatal("missing injected encoding failure")
	}
	loaded, err := Query[counter](db, "test_counter").Get(t.Context(), id)
	if err != nil || loaded.Exists {
		t.Fatal("state escaped realtime rollback")
	}
	db.realtime = encoder
	first, err := Command[counter](db, "test_counter").Execute(t.Context(), m, increment)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := Command[counter](db, "test_counter").Execute(t.Context(), m, func(a.Loaded[counter]) (a.Mutation[counter], error) {
		t.Fatal("replayed command ran again")
		return a.Mutation[counter]{}, nil
	})
	if err != nil || repeated != first {
		t.Fatal("recorded outcome changed")
	}
	var publications int
	if err := db.pool.QueryRow(t.Context(), `SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1`, id).Scan(&publications); err != nil || publications != 1 {
		t.Fatal("realtime intent is not singular")
	}
	row, ok, err := db.ClaimRealtime(t.Context())
	if err != nil || !ok {
		t.Fatalf("claim: %v", err)
	}
	if _, err := owner.Exec(t.Context(), `UPDATE cafe.realtime_dispatches SET lease_until=clock_timestamp()-interval '1 second' WHERE event_id=$1`, row.ID); err != nil {
		t.Fatal(err)
	}
	reclaimed, ok, err := db.ClaimRealtime(t.Context())
	if err != nil || !ok || reclaimed.ID != row.ID {
		t.Fatal("expired dispatch was not reclaimed")
	}
	if err := db.FinishRealtime(t.Context(), row, nil); err != nil {
		t.Fatal(err)
	}
	var completed bool
	if err := db.pool.QueryRow(t.Context(), `SELECT completed_at IS NOT NULL FROM cafe.realtime_dispatches WHERE event_id=$1`, row.ID).Scan(&completed); err != nil || completed {
		t.Fatal("stale worker completed a reclaimed lease")
	}
	if err := db.FinishRealtime(t.Context(), reclaimed, nil); err != nil {
		t.Fatal(err)
	}
	if err := db.pool.QueryRow(t.Context(), `SELECT completed_at IS NOT NULL FROM cafe.realtime_dispatches WHERE event_id=$1`, row.ID).Scan(&completed); err != nil || !completed {
		t.Fatal("current lease did not complete")
	}
}
