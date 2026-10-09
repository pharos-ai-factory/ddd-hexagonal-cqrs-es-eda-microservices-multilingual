//go:build integration

package postgres

import (
	"bytes"
	"encoding/json"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"testing"
	"time"
)

func TestCommandReplyAtomicityRetryAndLeaseFencing(t *testing.T) {
	db, pool := fixtureDB(t), admin(t)
	id := NewID()
	clean(t, pool, id)
	store := snapshotDecisions[counter](db, "test_counter")
	m := metadata(id, 0)
	intent := &ReplyIntent{ID: NewID(), Encode: func(a.Outcome) ([]byte, error) { return nil, fmt.Errorf("reply encoding failed") }}
	ctx := WithReply(t.Context(), intent)
	if _, err := store.Execute(ctx, m, increment); err == nil || intent.Committed {
		t.Fatal("reply encoding failure did not roll back")
	}
	loaded, err := NewSnapshotReadRepository[counter](db, "test_counter").Get(t.Context(), id)
	if err != nil || loaded.Exists {
		t.Fatal("root escaped reply rollback")
	}
	var n int
	if err := db.pool.QueryRow(t.Context(), `SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1`, id).Scan(&n); err != nil || n != 0 {
		t.Fatal("receipt escaped reply rollback")
	}
	intent.Encode = func(value a.Outcome) ([]byte, error) { return json.Marshal(value) }
	first, err := store.Execute(ctx, m, increment)
	if err != nil || !intent.Committed {
		t.Fatalf("commit: %v", err)
	}
	second := &ReplyIntent{ID: NewID(), Encode: intent.Encode}
	repeated, err := store.Execute(WithReply(t.Context(), second), m, func(a.Loaded[counter]) (a.Mutation[counter], error) {
		t.Fatal("retry ran decision")
		return a.Mutation[counter]{}, nil
	})
	if err != nil || repeated != first || !second.Committed {
		t.Fatal("retry did not persist recovered reply")
	}
	for _, request := range []*ReplyIntent{intent, second} {
		t.Cleanup(func() {
			pool.Exec(t.Context(), `DELETE FROM cafe.command_reply_dispatches WHERE event_id=$1`, request.ID)
			pool.Exec(t.Context(), `DELETE FROM cafe.command_replies WHERE id=$1`, request.ID)
		})
		var body []byte
		db.pool.QueryRow(t.Context(), `SELECT body FROM cafe.command_replies WHERE id=$1`, request.ID).Scan(&body)
		expected, _ := json.Marshal(first)
		if !bytes.Equal(body, expected) {
			t.Fatal("reply bytes changed")
		}
	}
	row, found, err := db.ClaimReply(t.Context())
	if err != nil || !found {
		t.Fatal("reply was not recoverable")
	}
	// Abandon this committed claim, as a process dying before publication would.
	// Keep real lease and expiry timestamps: shortening only the lease hides expiry bugs.
	if _, err = pool.Exec(t.Context(), `UPDATE cafe.command_reply_dispatches SET available_at=clock_timestamp()+interval '1 hour' WHERE event_id<>$1`, row.EventID); err != nil {
		t.Fatal(err)
	}
	var reclaimed ReplyDispatch
	deadline := time.Now().Add(11 * time.Second)
	for {
		reclaimed, found, err = db.ClaimReply(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if found {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("abandoned reply lease did not recover within the caller deadline")
		}
		time.Sleep(50 * time.Millisecond)
	}
	if reclaimed.EventID != row.EventID || reclaimed.Expired || !bytes.Equal(row.Body, reclaimed.Body) {
		t.Fatal("crash recovery lost the original live reply")
	}
	if err = db.FinishReply(t.Context(), row, ""); err != nil {
		t.Fatal(err)
	}
	var done bool
	db.pool.QueryRow(t.Context(), `SELECT completed_at IS NOT NULL FROM cafe.command_reply_dispatches WHERE event_id=$1`, row.EventID).Scan(&done)
	if done {
		t.Fatal("stale publisher completed a reclaimed reply")
	}
	if err = db.FinishReply(t.Context(), reclaimed, ""); err != nil {
		t.Fatal(err)
	}
	db.pool.QueryRow(t.Context(), `SELECT completed_at IS NOT NULL FROM cafe.command_reply_dispatches WHERE event_id=$1`, row.EventID).Scan(&done)
	if !done {
		t.Fatal("confirmed reply did not complete")
	}
	if _, err = db.pool.Exec(t.Context(), `UPDATE cafe.command_replies SET body=body WHERE id=$1`, row.EventID); err == nil {
		t.Fatal("runtime rewrote immutable reply bytes")
	}
}
