// Package command owns the infrastructure boundary around plain feature handlers.
package command

import (
	"context"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Result contains internal commit material for the transaction coordinator.
type Result struct {
	Status       string
	Publications []a.Publication
	Realtime     []a.RealtimeMessage
}

// Transaction coordinates one aggregate in one context database.
type Transaction[A any] interface {
	Execute(context.Context, a.Metadata, func(a.WriteRepository[A]) (Result, error)) (a.Outcome, error)
}

// FeatureHandler exposes business behaviour with a narrow target context.
type FeatureHandler[C any] interface {
	Execute(context.Context, a.CommandContext, C) (a.CommandResult, error)
}

// Executor is the typed transport-facing entry point with durable outcome semantics.
type Executor[C any] struct {
	run func(context.Context, a.Metadata, C) (a.Outcome, error)
}

func (e Executor[C]) Execute(ctx context.Context, m a.Metadata, c C) (a.Outcome, error) {
	if e.run == nil {
		return a.Outcome{}, fmt.Errorf("command executor is not registered")
	}
	return e.run(ctx, m, c)
}

type executor[A, C any, H FeatureHandler[C]] struct {
	transaction Transaction[A]
	factory     func(a.WriteRepository[A], a.RealtimeBusPort[a.RealtimeMessage]) H
}

// Bind constructs a fresh handler with its invocation's repository inside the transaction.
func Bind[A, C any, H FeatureHandler[C]](transaction Transaction[A], factory func(a.WriteRepository[A], a.RealtimeBusPort[a.RealtimeMessage]) H) Executor[C] {
	bound := &executor[A, C, H]{transaction, factory}
	return Executor[C]{run: bound.Execute}
}
func (e *executor[A, C, H]) Execute(ctx context.Context, m a.Metadata, c C) (a.Outcome, error) {
	return e.transaction.Execute(ctx, m, func(repository a.WriteRepository[A]) (Result, error) {
		bus := &RealtimeBuffer{active: true}
		defer bus.Close()
		result, err := e.factory(repository, bus).Execute(ctx, a.CommandContext{Target: m.AggregateID, ActorID: m.ActorID}, c)
		return Result{Status: result.Status, Realtime: bus.Events}, err
	})
}

// EventRecordingRepository stages publications from the saved aggregate's private facts.
type EventRecordingRepository[A any] struct {
	Source       a.WriteRepository[A]
	Map          func(A) []a.Publication
	Publications []a.Publication
}

func (r *EventRecordingRepository[A]) Get(ctx context.Context, id string) (a.Loaded[A], error) {
	return r.Source.Get(ctx, id)
}
func (r *EventRecordingRepository[A]) Save(ctx context.Context, aggregate A) error {
	var publications []a.Publication
	if r.Map != nil {
		publications = r.Map(aggregate)
	}
	r.Publications = nil
	var bus a.DomainEventBusPort[a.Publication] = r
	for _, publication := range publications {
		if err := bus.Publish(ctx, publication); err != nil {
			return err
		}
	}
	if err := r.Source.Save(ctx, aggregate); err != nil {
		return err
	}
	return nil
}

// Publish stages a private fact for the transaction; public contracts require their own subscriber.
func (r *EventRecordingRepository[A]) Publish(_ context.Context, publication a.Publication) error {
	if publication.Visibility != a.Private {
		return fmt.Errorf("aggregate saves may stage only private domain events")
	}
	r.Publications = append(r.Publications, publication)
	return nil
}
