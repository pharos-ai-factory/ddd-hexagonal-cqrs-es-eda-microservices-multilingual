//go:build integration

package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestSecondAggregateWriteIsRejectedByDatabase(t *testing.T) {
	db := fixtureDB(t)
	owner := admin(t)
	first, second := NewID(), NewID()
	clean(t, owner, first)
	clean(t, owner, second)
	tx, err := db.pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(t.Context(), `SELECT set_config('cafe.command_target',$1,true)`, "test_counter:"+first)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(t.Context(), `INSERT INTO cafe.aggregates(kind,id,version,state) VALUES('test_counter',$1,1,'{}')`, first)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(t.Context(), `SELECT set_config('cafe.command_target',$1,true)`, "test_counter:"+second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(t.Context(), `INSERT INTO cafe.aggregates(kind,id,version,state) VALUES('test_counter',$1,1,'{}')`, second)
	if err == nil {
		t.Fatal("a second aggregate entered the same transaction")
	}
	_ = tx.Rollback(t.Context())
	loaded, err := NewSnapshotReadRepository[counter](db, "test_counter").Get(t.Context(), first)
	if err != nil || loaded.Exists {
		t.Fatal("first aggregate did not roll back with the rejected second write")
	}
}
func TestRuntimeCannotRewriteEventsOrAdministerSchema(t *testing.T) {
	db := fixtureDB(t)
	for _, statement := range []string{`UPDATE cafe.outbox_events SET event_name='tampered'`, `DELETE FROM cafe.outbox_events`, `CREATE TABLE cafe.unauthorised(id int)`} {
		if _, err := db.pool.Exec(t.Context(), statement); err == nil {
			t.Fatalf("runtime accepted: %s", statement)
		}
	}
	parsed, err := url.Parse(os.Getenv("ORDERING_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/cafe_loyalty"
	foreign, err := pgxpool.New(t.Context(), parsed.String())
	if err != nil {
		return
	}
	defer foreign.Close()
	if err = foreign.Ping(t.Context()); err == nil {
		t.Fatal("Ordering connected to Loyalty's database")
	}
}
func TestDispatchLeaseFencesExpiredPublisher(t *testing.T) {
	db := fixtureDB(t)
	owner := admin(t)
	id := NewID()
	clean(t, owner, id)
	_, err := snapshotDecisions[counter](db, "test_counter").Execute(t.Context(), metadata(id, 0), func(a.Loaded[counter]) (a.Mutation[counter], error) {
		return a.Changed(counter{1}, "active", a.Publication{Name: "test.created", Visibility: a.Private}), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	first, found, err := db.Claim(t.Context(), time.Minute)
	if err != nil || !found {
		t.Fatalf("claim failed: %v", err)
	}
	if _, err = owner.Exec(t.Context(), `UPDATE cafe.dispatches SET lease_until=clock_timestamp()-interval '1 second' WHERE event_id=$1`, first.EventID); err != nil {
		t.Fatal(err)
	}
	second, found, err := db.Claim(t.Context(), time.Minute)
	if err != nil || !found || second.EventID != first.EventID || second.Generation <= first.Generation {
		t.Fatalf("reclaim failed %+v %v", second, err)
	}
	if done, err := db.Complete(t.Context(), first); err != nil || done {
		t.Fatal("expired owner completed a reclaimed dispatch")
	}
	if done, err := db.Retry(t.Context(), first, "stale retry"); err != nil || done {
		t.Fatal("expired owner rescheduled a reclaimed dispatch")
	}
	if done, err := db.Complete(t.Context(), second); err != nil || !done {
		t.Fatal("current lease could not complete")
	}
}
