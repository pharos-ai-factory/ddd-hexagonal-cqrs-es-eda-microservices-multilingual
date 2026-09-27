package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type Queries[S any] struct {
	db   *Database
	kind string
}

func Query[S any](db *Database, kind string) *Queries[S] { return &Queries[S]{db: db, kind: kind} }
func (q *Queries[S]) Get(ctx context.Context, id string) (a.Loaded[S], error) {
	var result a.Loaded[S]
	var data []byte
	err := q.db.pool.QueryRow(ctx, `SELECT version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2`, q.kind, id).Scan(&result.Version, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	result.Exists = true
	err = decodeStored(data, &result.State)
	if err == nil {
		err = checkRootIdentity(data, id)
	}
	return result, err
}
func (q *Queries[S]) List(ctx context.Context) ([]a.Loaded[S], error) {
	rows, err := q.db.pool.Query(ctx, `SELECT id,version,state FROM cafe.aggregates WHERE kind=$1 ORDER BY id LIMIT 100`, q.kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []a.Loaded[S]{}
	for rows.Next() {
		var item a.Loaded[S]
		var data []byte
		var id string
		if err = rows.Scan(&id, &item.Version, &data); err != nil {
			return nil, err
		}
		item.Exists = true
		if err = decodeStored(data, &item.State); err != nil {
			return nil, err
		}
		if err = checkRootIdentity(data, id); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
