//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"strings"
	"testing"
)

func TestExplicitRealtimeTransactionsAndSubscriberRecovery(t *testing.T) {
	db := fixtureDB(t)
	encoder := func(_, _, _, _ string, _ uint64, event a.RealtimeMessage) ([]byte, error) { return json.Marshal(event) }
	if err := db.EnableRealtime(t.Context(), encoder); err != nil {
		t.Fatal(err)
	}
	target := NewID()
	transaction := NewAggregateTransaction[counter, counter](db, "test_counter", func(r a.WriteRepository[counter]) a.WriteRepository[counter] { return r })
	run := func(value int, publish bool) func(a.WriteRepository[counter]) (execution.Result, error) {
		return func(r a.WriteRepository[counter]) (execution.Result, error) {
			err := r.Save(t.Context(), counter{value})
			result := execution.Result{Status: "changed"}
			if publish {
				result.Realtime = []a.RealtimeMessage{a.SnapshotMessage(a.ContextTarget(db.owner), counter{value})}
			}
			return result, err
		}
	}
	count := func() int {
		t.Helper()
		var n int
		if err := db.pool.QueryRow(t.Context(), "SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1", target).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if _, err := transaction.Execute(t.Context(), metadata(target, 0), run(1, false)); err != nil {
		t.Fatal(err)
	}
	if count() != 0 {
		t.Fatal("saving implicitly published realtime")
	}
	m := metadata(target, 1)
	db.realtime = func(_, _, _, _ string, _ uint64, _ a.RealtimeMessage) ([]byte, error) {
		return nil, errors.New("encoding failure")
	}
	if _, err := transaction.Execute(t.Context(), m, run(2, true)); err == nil {
		t.Fatal("missing encoding failure")
	}
	loaded, err := NewSnapshotReadRepository[counter](db, "test_counter").Get(t.Context(), target)
	if err != nil || loaded.Version != 1 || count() != 0 {
		t.Fatal("failed publication escaped rollback")
	}
	db.realtime = encoder
	if _, err := transaction.Execute(t.Context(), m, run(2, true)); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.Execute(t.Context(), m, func(a.WriteRepository[counter]) (execution.Result, error) {
		t.Fatal("duplicate command executed")
		return execution.Result{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if count() != 1 {
		t.Fatal("command publication duplicated")
	}
	source := a.Metadata{ID: NewID(), AggregateID: target, Consumer: "ordering.realtime-test", SourceEventID: NewID(), SourceHash: strings.Repeat("a", 64), CorrelationID: NewID()}
	bus := NewRealtimeMessageBus(db, source.Consumer, "test_counter")
	event := a.SnapshotMessage(a.ContextTarget(db.owner), counter{2})
	event.Source, event.Revision = &source, loaded.Version+1
	for range 2 {
		if err := bus.Publish(t.Context(), event); err != nil {
			t.Fatal(err)
		}
	}
	if count() != 2 {
		t.Fatal("subscription redelivery duplicated publication")
	}
	source.SourceHash = strings.Repeat("b", 64)
	if err := bus.Publish(t.Context(), event); err == nil {
		t.Fatal("conflicting source bytes accepted")
	}

	// A recorded rejection can publish without creating an aggregate.
	missing := metadata(NewID(), 0)
	missing.ActorID = NewID()
	rejected := d.Violation{Code: "not_found", Message: "The resource does not exist"}
	outcome, err := transaction.Execute(t.Context(), missing, func(a.WriteRepository[counter]) (execution.Result, error) {
		return execution.Result{Realtime: []a.RealtimeMessage{a.RejectionMessage(a.PrincipalTarget(missing.ActorID), rejected)}}, &rejected
	})
	if err != nil || outcome.Rejection == nil || outcome.Version != 0 {
		t.Fatalf("rejection publication: %+v %v", outcome, err)
	}
	var n int
	if err = db.pool.QueryRow(context.Background(), "SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id=$1 AND message_kind='operation_rejected'", missing.AggregateID).Scan(&n); err != nil || n != 1 {
		t.Fatal("missing durable rejection")
	}
}
