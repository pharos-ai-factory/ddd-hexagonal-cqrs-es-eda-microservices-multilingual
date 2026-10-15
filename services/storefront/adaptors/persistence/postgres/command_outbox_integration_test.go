//go:build integration

package postgres

import (
	"bytes"
	"context"
	"fmt"
	ownercodec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCommandAcceptanceRollbackDuplicateIdentityAndFencedRecovery(t *testing.T) {
	db := fixtureDB(t)
	owner := admin(t)
	ctx := t.Context()
	codec := ownercodec.RecordCollectionCodec{}
	bus := NewCommandOutbox[orderingcontracts.RecordCollectionCommand](db, codec)
	source, target, customer := NewID(), NewID(), NewID()
	m := a.Metadata{ID: DerivedID(codec.Consumer(), source), Name: codec.Consumer(), Consumer: codec.Consumer(), SourceEventID: source, CausationID: source, AggregateID: target, CorrelationID: NewID(), SourceHash: strings.Repeat("a", 64), Input: map[string]any{"customerId": customer, "orderId": target}}
	command := orderingcontracts.RecordCollectionCommand{CustomerID: customer}
	t.Cleanup(func() {
		_, _ = owner.Exec(context.Background(), `DELETE FROM cafe.internal_command_dispatches WHERE event_id=$1`, m.ID)
		_, _ = owner.Exec(context.Background(), `DELETE FROM cafe.internal_commands WHERE id=$1`, m.ID)
	})
	count := func(table string) int {
		t.Helper()
		var n int
		if err := owner.QueryRow(ctx, "SELECT count(*) FROM cafe."+table+" WHERE "+map[string]string{"internal_commands": "id", "internal_command_dispatches": "event_id"}[table]+"=$1", m.ID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if _, err := owner.Exec(ctx, fmt.Sprintf(`ALTER TABLE cafe.internal_command_dispatches ADD CONSTRAINT reject_command_probe CHECK(event_id<>'%s') NOT VALID`, m.ID)); err != nil {
		t.Fatal(err)
	}
	_, failure := bus.Enqueue(ctx, m, command)
	if _, err := owner.Exec(ctx, `ALTER TABLE cafe.internal_command_dispatches DROP CONSTRAINT reject_command_probe`); err != nil {
		t.Fatal(err)
	}
	if failure == nil || count("internal_commands") != 0 || count("internal_command_dispatches") != 0 {
		t.Fatal("failed acceptance escaped rollback")
	}
	var group sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		group.Go(func() {
			out, err := bus.Enqueue(ctx, m, command)
			if err == nil && out.Status != "queued" {
				err = fmt.Errorf("unexpected acceptance: %+v", out)
			}
			results <- err
		})
	}
	group.Wait()
	for range 2 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	if count("internal_commands") != 1 || count("internal_command_dispatches") != 1 {
		t.Fatal("duplicate acceptance")
	}
	var original []byte
	if err := owner.QueryRow(ctx, `SELECT body FROM cafe.internal_commands WHERE id=$1`, m.ID).Scan(&original); err != nil {
		t.Fatal(err)
	}
	// A later application mapping cannot rewrite already accepted command bytes.
	command.CustomerID = NewID()
	if _, err := bus.Enqueue(ctx, m, command); err != nil {
		t.Fatal(err)
	}
	var replay []byte
	_ = owner.QueryRow(ctx, `SELECT body FROM cafe.internal_commands WHERE id=$1`, m.ID).Scan(&replay)
	if !bytes.Equal(original, replay) {
		t.Fatal("accepted command was remapped")
	}
	for _, change := range []func(*a.Metadata){func(m *a.Metadata) { m.SourceHash = strings.Repeat("b", 64) }, func(m *a.Metadata) { m.AggregateID = NewID() }} {
		invalid := m
		change(&invalid)
		if _, err := bus.Enqueue(ctx, invalid, command); err == nil {
			t.Fatal("conflicting receiving evidence accepted")
		}
	}
	if _, err := db.pool.Exec(ctx, `UPDATE cafe.internal_commands SET body=$2 WHERE id=$1`, m.ID, []byte("changed")); err == nil {
		t.Fatal("runtime rewrote immutable bytes")
	}
	// Select this dispatch deterministically while other packages use the same fixture database.
	if _, err := owner.Exec(ctx, `UPDATE cafe.internal_command_dispatches SET available_at='1970-01-01' WHERE event_id=$1`, m.ID); err != nil {
		t.Fatal(err)
	}
	first, found, err := db.ClaimCommand(ctx, time.Minute)
	if err != nil || !found || first.EventID != m.ID {
		t.Fatalf("claim: %+v %v", first, err)
	}
	if _, err = owner.Exec(ctx, `UPDATE cafe.internal_command_dispatches SET lease_until=clock_timestamp()-interval '1 second' WHERE event_id=$1`, m.ID); err != nil {
		t.Fatal(err)
	}
	second, found, err := db.ClaimCommand(ctx, time.Minute)
	if err != nil || !found || second.Generation <= first.Generation || !bytes.Equal(first.Body, second.Body) {
		t.Fatalf("reclaim: %+v %v", second, err)
	}
	if done, err := db.CompleteCommand(ctx, first); err != nil || done {
		t.Fatal("expired publisher completed reclaimed command")
	}
	if done, err := db.RetryCommand(ctx, first, "stale"); err != nil || done {
		t.Fatal("expired publisher changed retry state")
	}
	if done, err := db.CompleteCommand(ctx, second); err != nil || !done {
		t.Fatal("current publisher could not complete")
	}
}
