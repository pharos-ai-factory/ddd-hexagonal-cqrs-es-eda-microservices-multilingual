//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"net/url"
	"os"
	"sync"
	"testing"
)

type counter struct {
	Value int `json:"value"`
}

func fixtureDB(t *testing.T) *ContextDatabase {
	t.Helper()
	raw := os.Getenv("ORDERING_DATABASE_URL")
	if raw == "" {
		t.Fatal("integration lane requires ORDERING_DATABASE_URL")
	}
	db, err := Open(t.Context(), raw, "ordering", func(m a.Message) ([]byte, error) { return json.Marshal(m) })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return db
}
func admin(t *testing.T) *pgxpool.Pool {
	t.Helper()
	raw := os.Getenv("DATABASE_ADMIN_URL")
	if raw == "" {
		t.Fatal("DATABASE_ADMIN_URL required")
	}
	pool, err := pgxpool.New(t.Context(), raw)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}
func metadata(id string, version uint64) a.Metadata {
	return a.Metadata{ID: NewID(), AggregateID: id, Name: "test.increment", ExpectedVersion: a.Expected(version), CorrelationID: NewID(), Input: counter{1}}
}
func clean(t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		for _, statement := range []string{
			`DELETE FROM cafe.realtime_dispatches WHERE event_id IN(SELECT id FROM cafe.realtime_publications WHERE aggregate_id=$1)`,
			`DELETE FROM cafe.realtime_publications WHERE aggregate_id=$1`,
			`DELETE FROM cafe.dispatches WHERE event_id IN(SELECT id FROM cafe.outbox_events WHERE aggregate_id=$1)`,
			`DELETE FROM cafe.outbox_events WHERE aggregate_id=$1`,
			`DELETE FROM cafe.command_receipts WHERE aggregate_id=$1`,
			`DELETE FROM cafe.aggregates WHERE id=$1`,
		} {
			if _, err := pool.Exec(ctx, statement, id); err != nil {
				t.Errorf("fixture cleanup: %v", err)
			}
		}
	})
}
func increment(s a.Loaded[counter]) (a.Mutation[counter], error) {
	return a.Changed(counter{s.State.Value + 1}, "active"), nil
}
func TestAtomicReceiptAndOptimisticConcurrency(t *testing.T) {
	db := fixtureDB(t)
	owner := admin(t)
	id := NewID()
	clean(t, owner, id)
	store := snapshotDecisions[counter](db, "test_counter")
	m := metadata(id, 0)
	decide := func(s a.Loaded[counter]) (a.Mutation[counter], error) {
		return a.Changed(counter{1}, "active", a.Publication{Name: "test.created", Visibility: a.Private, Payload: counter{1}}), nil
	}
	first, err := store.Execute(t.Context(), m, decide)
	if err != nil || first.Rejection != nil {
		t.Fatalf("initial transaction: %+v %v", first, err)
	}
	repeated, err := store.Execute(t.Context(), m, func(a.Loaded[counter]) (a.Mutation[counter], error) {
		t.Fatal("replay invoked business behaviour")
		return a.Mutation[counter]{}, nil
	})
	if err != nil || repeated != first {
		t.Fatalf("recorded outcome changed: %+v %v", repeated, err)
	}
	changed := m
	changed.Input = counter{2}
	conflict, err := store.Execute(t.Context(), changed, decide)
	if err != nil || conflict.Rejection == nil || conflict.Rejection.Code != "idempotency_conflict" {
		t.Fatalf("input reuse: %+v %v", conflict, err)
	}
	var events, receipts int
	_ = db.pool.QueryRow(t.Context(), `SELECT count(*) FROM cafe.outbox_events WHERE aggregate_id=$1`, id).Scan(&events)
	_ = db.pool.QueryRow(t.Context(), `SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1`, id).Scan(&receipts)
	if events != 1 || receipts != 1 {
		t.Fatalf("atomic records events=%d receipts=%d", events, receipts)
	}
	var wait sync.WaitGroup
	results := make(chan a.Outcome, 2)
	errors := make(chan error, 2)
	for range 2 {
		wait.Go(func() {
			result, err := store.Execute(t.Context(), metadata(id, 1), increment)
			results <- result
			errors <- err
		})
	}
	wait.Wait()
	close(results)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	success, stale := 0, 0
	for result := range results {
		if result.Rejection == nil {
			success++
		} else if result.Rejection.Code == "version_conflict" {
			stale++
		}
	}
	if success != 1 || stale != 1 {
		t.Fatalf("concurrent decisions success=%d stale=%d", success, stale)
	}
	loaded, err := NewSnapshotReadRepository[counter](db, "test_counter").Get(t.Context(), id)
	if err != nil || loaded.Version != 2 || loaded.State.Value != 2 {
		t.Fatalf("invalid final counter %+v %v", loaded, err)
	}
}
func TestOutboxFailureRollsBackAggregateAndReceipt(t *testing.T) {
	db := fixtureDB(t)
	owner := admin(t)
	id := NewID()
	clean(t, owner, id)
	broken := *db
	broken.encode = func(a.Message) ([]byte, error) {
		return nil, fmt.Errorf("injected publication failure after aggregate SQL")
	}
	_, err := snapshotDecisions[counter](&broken, "test_counter").Execute(t.Context(), metadata(id, 0), func(a.Loaded[counter]) (a.Mutation[counter], error) {
		return a.Changed(counter{1}, "active", a.Publication{Name: "test.created", Visibility: a.Private}), nil
	})
	if err == nil {
		t.Fatal("injected failure was lost")
	}
	loaded, err := NewSnapshotReadRepository[counter](db, "test_counter").Get(t.Context(), id)
	if err != nil || loaded.Exists {
		t.Fatal("aggregate escaped rollback")
	}
	var count int
	_ = db.pool.QueryRow(t.Context(), `SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1`, id).Scan(&count)
	if count != 0 {
		t.Fatal("receipt escaped rollback")
	}
}
func TestDatabaseConnectionLossRollsBackUncommittedState(t *testing.T) {
	owner := admin(t)
	normal := fixtureDB(t)
	id := NewID()
	clean(t, owner, id)
	parsed, err := url.Parse(os.Getenv("ORDERING_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	label := "failure-" + NewID()
	query.Set("application_name", label)
	parsed.RawQuery = query.Encode()
	encoder := func(message a.Message) ([]byte, error) {
		_, err := owner.Exec(t.Context(), `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE application_name=$1`, label)
		if err != nil {
			return nil, err
		}
		return json.Marshal(message)
	}
	db, err := Open(t.Context(), parsed.String(), "ordering", encoder)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = snapshotDecisions[counter](db, "test_counter").Execute(t.Context(), metadata(id, 0), func(a.Loaded[counter]) (a.Mutation[counter], error) {
		return a.Changed(counter{1}, "active", a.Publication{Name: "test.created", Visibility: a.Private}), nil
	})
	if err == nil {
		t.Fatal("connection termination did not fail the transaction")
	}
	loaded, err := NewSnapshotReadRepository[counter](normal, "test_counter").Get(t.Context(), id)
	if err != nil || loaded.Exists {
		t.Fatalf("state survived uncommitted connection loss: %+v %v", loaded, err)
	}
}
