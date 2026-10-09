package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// SnapshotReadRepository reads one aggregate kind and validates stored identities.
type SnapshotReadRepository[S any] struct {
	db   *ContextDatabase
	kind string
}

// NewSnapshotReadRepository constructs a snapshot reader for one aggregate kind.
func NewSnapshotReadRepository[S any](db *ContextDatabase, kind string) *SnapshotReadRepository[S] {
	return &SnapshotReadRepository[S]{db: db, kind: kind}
}
func (q *SnapshotReadRepository[S]) Get(ctx context.Context, id string) (a.Loaded[S], error) {
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
func (q *SnapshotReadRepository[S]) List(ctx context.Context) ([]a.Loaded[S], error) {
	result, _, err := q.read(ctx, nil)
	return result, err
}
func (q *SnapshotReadRepository[S]) Page(ctx context.Context, request a.PageRequest) (a.Page[S], error) {
	if request.Limit < 1 || request.Limit > 100 {
		return a.Page[S]{}, errors.New("invalid page size")
	}
	items, ids, err := q.read(ctx, &request)
	page := a.Page[S]{Items: items}
	if len(items) > request.Limit {
		page.Items = items[:request.Limit]
		page.NextID = ids[request.Limit-1]
	}
	return page, err
}
func (q *SnapshotReadRepository[S]) read(ctx context.Context, request *a.PageRequest) ([]a.Loaded[S], []string, error) {
	sql := `SELECT id,version,state FROM cafe.aggregates WHERE kind=$1`
	args := []any{q.kind}
	if request != nil {
		sql += ` AND ($2::uuid IS NULL OR id>$2::uuid) ORDER BY id LIMIT $3`
		var after any
		if request.After != "" {
			after = request.After
		}
		args = append(args, after, request.Limit+1)
	} else {
		sql += ` ORDER BY id`
	}
	rows, err := q.db.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	result := []a.Loaded[S]{}
	ids := []string{}
	for rows.Next() {
		var item a.Loaded[S]
		var data []byte
		var id string
		if err = rows.Scan(&id, &item.Version, &data); err != nil {
			return nil, nil, err
		}
		item.Exists = true
		if err = decodeStored(data, &item.State); err != nil {
			return nil, nil, err
		}
		if err = checkRootIdentity(data, id); err != nil {
			return nil, nil, err
		}
		result = append(result, item)
		ids = append(ids, id)
	}
	return result, ids, rows.Err()
}
