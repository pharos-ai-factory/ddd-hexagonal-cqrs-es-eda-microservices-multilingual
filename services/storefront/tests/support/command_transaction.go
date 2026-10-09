package support

import (
	"context"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
)

// ForAggregate adapts a decision probe to an aggregate-aware command transaction.
func ForAggregate[S, A any](probe *CommandProbe[S], restore func(S) (A, error), snapshot func(A) S, publications func(A) []a.Publication) execution.Transaction[A] {
	return &aggregateProbe[S, A]{probe, restore, snapshot, publications}
}

type aggregateProbe[S, A any] struct {
	probe        *CommandProbe[S]
	restore      func(S) (A, error)
	snapshot     func(A) S
	publications func(A) []a.Publication
}

func (p *aggregateProbe[S, A]) Execute(ctx context.Context, m a.Metadata, work func(a.WriteRepository[A]) (execution.Result, error)) (a.Outcome, error) {
	return p.probe.Execute(ctx, m, func(loaded a.Loaded[S]) (a.Mutation[S], error) {
		repo := &probeRepository[S, A]{target: m.AggregateID, loaded: loaded, restore: p.restore, snapshot: p.snapshot}
		recording := &execution.EventRecordingRepository[A]{Source: repo, Map: p.publications}
		result, err := work(recording)
		result.Publications = append(result.Publications, recording.Publications...)
		return a.Mutation[S]{State: repo.saved, Changed: repo.changed, Status: result.Status, Publications: result.Publications}, err
	})
}

type probeRepository[S, A any] struct {
	target   string
	loaded   a.Loaded[S]
	restore  func(S) (A, error)
	snapshot func(A) S
	saved    S
	changed  bool
}

func (p *probeRepository[S, A]) Get(_ context.Context, id string) (a.Loaded[A], error) {
	if id != p.target {
		return a.Loaded[A]{}, fmt.Errorf("unexpected aggregate target")
	}
	if !p.loaded.Exists {
		return a.Loaded[A]{}, nil
	}
	state, err := p.restore(clone(p.loaded.State))
	return a.Loaded[A]{Exists: true, Version: p.loaded.Version, State: state}, err
}
func (p *probeRepository[S, A]) Save(_ context.Context, aggregate A) error {
	p.saved = clone(p.snapshot(aggregate))
	p.changed = true
	return nil
}

// SnapshotDecisionFixture exercises the real transaction using technical snapshot fixtures.
type SnapshotDecisionFixture[S any] struct{ Transaction execution.Transaction[S] }

func (p SnapshotDecisionFixture[S]) Execute(ctx context.Context, m a.Metadata, decide func(a.Loaded[S]) (a.Mutation[S], error)) (a.Outcome, error) {
	return p.Transaction.Execute(ctx, m, func(repo a.WriteRepository[S]) (execution.Result, error) {
		loaded, err := repo.Get(ctx, m.AggregateID)
		if err != nil {
			return execution.Result{}, err
		}
		change, err := decide(loaded)
		if err != nil {
			return execution.Result{}, err
		}
		if change.Changed {
			if err = repo.Save(ctx, change.State); err != nil {
				return execution.Result{}, err
			}
		}
		return execution.Result{Status: change.Status, Publications: change.Publications}, nil
	})
}
