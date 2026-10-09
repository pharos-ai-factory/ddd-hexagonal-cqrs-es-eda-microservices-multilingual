package postgres

import (
	"context"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// MappedWriteRepository restores aggregates and stages snapshots through an owner mapping.
type MappedWriteRepository[S, A any] struct {
	Source   a.WriteRepository[S]
	Restore  func(S) (A, error)
	Snapshot func(A) S
}

func (r MappedWriteRepository[S, A]) Get(ctx context.Context, id string) (a.Loaded[A], error) {
	loaded, err := r.Source.Get(ctx, id)
	result := a.Loaded[A]{Exists: loaded.Exists, Version: loaded.Version}
	if err != nil || !loaded.Exists {
		return result, err
	}
	result.State, err = r.Restore(loaded.State)
	return result, err
}
func (r MappedWriteRepository[S, A]) Save(ctx context.Context, root A) error {
	return r.Source.Save(ctx, r.Snapshot(root))
}
