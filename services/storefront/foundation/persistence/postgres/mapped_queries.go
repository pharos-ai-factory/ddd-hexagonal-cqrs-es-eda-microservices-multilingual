package postgres

import (
	"context"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// MappedQueries selects application views from validated owner snapshots.
type MappedQueries[S, V any] struct {
	Source a.PagedQueryPort[S]
	View   func(S) (V, error)
}

func (q MappedQueries[S, V]) mapped(value a.Loaded[S]) (a.Loaded[V], error) {
	result := a.Loaded[V]{Exists: value.Exists, Version: value.Version}
	if !value.Exists {
		return result, nil
	}
	var err error
	result.State, err = q.View(value.State)
	return result, err
}
func (q MappedQueries[S, V]) Get(ctx context.Context, id string) (a.Loaded[V], error) {
	value, err := q.Source.Get(ctx, id)
	if err != nil {
		return a.Loaded[V]{}, err
	}
	return q.mapped(value)
}
func (q MappedQueries[S, V]) items(values []a.Loaded[S]) ([]a.Loaded[V], error) {
	result := make([]a.Loaded[V], 0, len(values))
	for _, value := range values {
		item, err := q.mapped(value)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}
func (q MappedQueries[S, V]) List(ctx context.Context) ([]a.Loaded[V], error) {
	values, err := q.Source.List(ctx)
	if err != nil {
		return nil, err
	}
	return q.items(values)
}
func (q MappedQueries[S, V]) Page(ctx context.Context, p a.PageRequest) (a.Page[V], error) {
	page, err := q.Source.Page(ctx, p)
	if err != nil {
		return a.Page[V]{}, err
	}
	items, err := q.items(page.Items)
	return a.Page[V]{Items: items, NextID: page.NextID}, err
}
