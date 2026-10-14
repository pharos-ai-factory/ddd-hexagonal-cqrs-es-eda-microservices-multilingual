package command

import (
	"context"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// BindProjection reads an immutable owner projection before opening the write transaction.
// This avoids taking a second pooled connection while holding the aggregate lock.
func BindProjection[A, C, S any, H FeatureHandler[C]](transaction Transaction[A], source a.ProjectionPort[S], key func(C) string,
	factory func(a.WriteRepository[A], a.ProjectionPort[S]) H) Executor[C] {
	return Executor[C]{run: func(ctx context.Context, m a.Metadata, c C) (a.Outcome, error) {
		identity := key(c)
		value, found, err := source.Find(ctx, identity)
		if err != nil {
			return a.Outcome{}, err
		}
		projection := preparedProjection[S]{identity, value, found}
		return Bind(transaction, func(repository a.WriteRepository[A]) H { return factory(repository, projection) }).Execute(ctx, m, c)
	}}
}

type preparedProjection[S any] struct {
	key   string
	value S
	found bool
}

func (p preparedProjection[S]) Find(_ context.Context, key string) (S, bool, error) {
	if key != p.key {
		var zero S
		return zero, false, fmt.Errorf("command requested an unprepared projection")
	}
	return p.value, p.found, nil
}
func (p preparedProjection[S]) Record(context.Context, a.Metadata, string, uint64, S) error {
	return fmt.Errorf("command projection is read-only")
}
