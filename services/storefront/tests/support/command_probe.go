// Package support provides decision probes for application scenarios.
// These probes do not implement receipts, locking or delivery guarantees.
package support

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

const Customer = "00000000-0000-4000-8000-000000000001"
const Drink = "00000000-0000-4000-8000-000000000002"
const Edition = "00000000-0000-4000-8000-000000000003"
const Order = "00000000-0000-4000-8000-000000000004"
const FirstLine = "00000000-0000-4000-8000-000000000005"
const SecondLine = "00000000-0000-4000-8000-000000000006"

func Metadata(id string) a.Metadata {
	return a.Metadata{ID: Customer, AggregateID: id, CorrelationID: Customer}
}

type CommandProbe[S any] struct {
	Loaded a.Loaded[S]
	Before a.Loaded[S]
	Last   a.Mutation[S]
	Result a.Outcome
}

func clone[S any](value S) S {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	var copy S
	if err = json.Unmarshal(data, &copy); err != nil {
		panic(err)
	}
	return copy
}

func (p *CommandProbe[S]) Execute(_ context.Context, m a.Metadata, decide func(a.Loaded[S]) (a.Mutation[S], error)) (a.Outcome, error) {
	p.Before = clone(p.Loaded)
	p.Last = a.Mutation[S]{}
	p.Result = a.Outcome{AggregateID: m.AggregateID, Version: p.Loaded.Version}
	mutation, err := decide(clone(p.Loaded))
	if err != nil {
		var rejection *d.Violation
		if errors.As(err, &rejection) {
			p.Result.Rejection = rejection
			return p.Result, nil
		}
		return p.Result, err
	}
	p.Last = mutation
	if mutation.Changed {
		p.Loaded = a.Loaded[S]{Exists: true, Version: p.Loaded.Version + 1, State: clone(mutation.State)}
	}
	p.Result.Version, p.Result.Status = p.Loaded.Version, mutation.Status
	return p.Result, nil
}

func (p *CommandProbe[S]) Unchanged() error {
	before, _ := json.Marshal(p.Before)
	after, _ := json.Marshal(p.Loaded)
	if string(before) != string(after) || len(p.Last.Publications) != 0 {
		return fmt.Errorf("rejected/no-op command changed state or outgoing publications")
	}
	return nil
}

func (p *CommandProbe[S]) Rejected(code string) error {
	if p.Result.Rejection == nil || p.Result.Rejection.Code != code {
		return fmt.Errorf("expected %s, got %+v", code, p.Result)
	}
	return nil
}

func (p *CommandProbe[S]) Succeeded() error {
	if p.Result.Rejection != nil {
		return fmt.Errorf("unexpected rejection: %+v", p.Result.Rejection)
	}
	return nil
}

type ProjectionProbe[S any] struct{ Values map[string]S }

func (p *ProjectionProbe[S]) Find(_ context.Context, key string) (S, bool, error) {
	value, found := p.Values[key]
	return clone(value), found, nil
}
func (p *ProjectionProbe[S]) Record(_ context.Context, _ a.Metadata, key string, _ uint64, value S) error {
	p.Values[key] = clone(value)
	return nil
}
