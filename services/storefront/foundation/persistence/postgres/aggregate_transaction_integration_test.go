//go:build integration

package postgres

import (
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	"testing"
)

type scopedState struct {
	ID     string `json:"id"`
	Values []int  `json:"values"`
}

func TestCommandRepositoryScopeLifetimeAndRejectedSave(t *testing.T) {
	db := fixtureDB(t)
	id := NewID()
	clean(t, admin(t), id)
	transaction := NewAggregateTransaction[scopedState, scopedState](db, "test_scope", func(repo a.WriteRepository[scopedState]) a.WriteRepository[scopedState] { return repo })
	queries := NewSnapshotReadRepository[scopedState](db, "test_scope")
	var escaped a.WriteRepository[scopedState]
	state := scopedState{ID: id, Values: []int{1}}
	first, err := transaction.Execute(t.Context(), metadata(id, 0), func(repo a.WriteRepository[scopedState]) (execution.Result, error) {
		escaped = repo
		if _, err := repo.Get(t.Context(), NewID()); err == nil {
			t.Fatal("foreign target allowed")
		}
		if err := repo.Save(t.Context(), scopedState{ID: NewID()}); err == nil {
			t.Fatal("foreign snapshot allowed")
		}
		if err := repo.Save(t.Context(), state); err != nil {
			return execution.Result{}, err
		}
		state.Values[0] = 9
		return execution.Result{Status: "active"}, nil
	})
	if err != nil || first.Version != 1 {
		t.Fatalf("create: %+v %v", first, err)
	}
	if _, err = escaped.Get(t.Context(), id); err == nil {
		t.Fatal("escaped read allowed")
	}
	if err = escaped.Save(t.Context(), state); err == nil {
		t.Fatal("escaped save allowed")
	}
	loaded, err := queries.Get(t.Context(), id)
	if err != nil || loaded.State.Values[0] != 1 {
		t.Fatalf("snapshot alias escaped staging: %+v %v", loaded, err)
	}
	rejectedMetadata := metadata(id, 1)
	rejected, err := transaction.Execute(t.Context(), rejectedMetadata, func(repo a.WriteRepository[scopedState]) (execution.Result, error) {
		escaped = repo
		if err := repo.Save(t.Context(), state); err != nil {
			return execution.Result{}, err
		}
		return execution.Result{}, d.Reject("fixture_rejection", "Reject after staging")
	})
	if err != nil || rejected.Rejection == nil || rejected.Version != 1 {
		t.Fatalf("rejection: %+v %v", rejected, err)
	}
	replay, err := transaction.Execute(t.Context(), rejectedMetadata, func(a.WriteRepository[scopedState]) (execution.Result, error) {
		t.Fatal("rejected replay ran again")
		return execution.Result{}, nil
	})
	if err != nil || replay.Rejection == nil || replay.Rejection.Code != "fixture_rejection" {
		t.Fatalf("replay: %+v %v", replay, err)
	}
	if err = escaped.Save(t.Context(), state); err == nil {
		t.Fatal("rejected repository escaped")
	}
	_, err = transaction.Execute(t.Context(), metadata(id, 1), func(repo a.WriteRepository[scopedState]) (execution.Result, error) {
		escaped = repo
		if err := repo.Save(t.Context(), state); err != nil {
			return execution.Result{}, err
		}
		return execution.Result{}, fmt.Errorf("infrastructure failure")
	})
	if err == nil {
		t.Fatal("infrastructure failure lost")
	}
	if _, err = escaped.Get(t.Context(), id); err == nil {
		t.Fatal("failed repository escaped")
	}
	loaded, err = queries.Get(t.Context(), id)
	if err != nil || loaded.Version != 1 || loaded.State.Values[0] != 1 {
		t.Fatalf("failed writes persisted: %+v %v", loaded, err)
	}
	var count int
	if err = db.pool.QueryRow(t.Context(), "SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id=$1", id).Scan(&count); err != nil || count != 2 {
		t.Fatalf("receipt count: %d %v", count, err)
	}
}
