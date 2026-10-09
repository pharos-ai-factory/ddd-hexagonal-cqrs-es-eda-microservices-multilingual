//go:build integration

package postgres

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

func TestConsumerIdentityCannotChangeTargetsConcurrently(t *testing.T) {
	db, owner := fixtureDB(t), admin(t)
	event, consumer := NewID(), "ordering.test-"+NewID()
	targets := []string{NewID(), NewID()}
	for _, target := range targets {
		clean(t, owner, target)
	}
	t.Cleanup(func() {
		_, err := owner.Exec(context.Background(), `DELETE FROM cafe.consumer_receipts WHERE consumer=$1`, consumer)
		if err != nil {
			t.Error(err)
		}
	})
	var decisions atomic.Int32
	var wait sync.WaitGroup
	results := make(chan error, 2)
	start := make(chan struct{})
	for _, target := range targets {
		wait.Go(func() {
			<-start
			m := metadata(target, 0)
			m.Consumer, m.SourceEventID, m.SourceHash = consumer, event, "same-wire-bytes"
			_, err := snapshotDecisions[counter](db, "test_counter").Execute(t.Context(), m, func(s a.Loaded[counter]) (a.Mutation[counter], error) {
				decisions.Add(1)
				return increment(s)
			})
			results <- err
		})
	}
	close(start)
	wait.Wait()
	close(results)
	failed := 0
	for err := range results {
		if err != nil {
			failed++
		}
	}
	if failed != 1 || decisions.Load() != 1 {
		t.Fatalf("conflicting delivery escaped receipt ownership: failures=%d decisions=%d", failed, decisions.Load())
	}
	count := 0
	for _, target := range targets {
		state, err := NewSnapshotReadRepository[counter](db, "test_counter").Get(t.Context(), target)
		if err != nil {
			t.Fatal(err)
		}
		if state.Exists {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("one delivery changed %d aggregates", count)
	}
}
