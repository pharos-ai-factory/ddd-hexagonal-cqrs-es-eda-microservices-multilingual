package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// SnapshotWriteRepository owns aggregate SQL inside a single-target command transaction.
type SnapshotWriteRepository[S any] struct {
	tx           pgx.Tx
	kind, target string
	loaded       a.Loaded[S]
	pending      *S
	active       bool
}

func newSnapshotWriteRepository[S any](ctx context.Context, tx pgx.Tx, kind, target string) (*SnapshotWriteRepository[S], error) {
	r := &SnapshotWriteRepository[S]{tx: tx, kind: kind, target: target, active: true}
	var data []byte
	err := tx.QueryRow(ctx, `SELECT version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2 FOR UPDATE`, kind, target).Scan(&r.loaded.Version, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}
	r.loaded.Exists = true
	if err = decodeStored(data, &r.loaded.State); err != nil {
		return nil, err
	}
	if err = checkRootIdentity(data, target); err != nil {
		return nil, err
	}
	return r, nil
}
func (r *SnapshotWriteRepository[S]) check(ctx context.Context) error {
	if !r.active {
		return fmt.Errorf("write repository used outside its command transaction")
	}
	return ctx.Err()
}
func (r *SnapshotWriteRepository[S]) Get(ctx context.Context, id string) (a.Loaded[S], error) {
	if err := r.check(ctx); err != nil {
		return a.Loaded[S]{}, err
	}
	if id != r.target {
		return a.Loaded[S]{}, fmt.Errorf("command transaction cannot access another aggregate")
	}
	value := r.loaded
	if r.pending != nil {
		value.Exists = true
		value.State = *r.pending
	}
	data, err := json.Marshal(value)
	if err != nil {
		return a.Loaded[S]{}, err
	}
	var copy a.Loaded[S]
	err = json.Unmarshal(data, &copy)
	return copy, err
}
func (r *SnapshotWriteRepository[S]) Save(ctx context.Context, state S) error {
	if err := r.check(ctx); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err = checkRootIdentity(data, r.target); err != nil {
		return err
	}
	var copy S
	if err = json.Unmarshal(data, &copy); err != nil {
		return err
	}
	r.pending = &copy
	return nil
}
func (r *SnapshotWriteRepository[S]) close() { r.active = false }
func (r *SnapshotWriteRepository[S]) flush(ctx context.Context) (uint64, []byte, error) {
	if r.pending == nil {
		return r.loaded.Version, nil, nil
	}
	data, err := json.Marshal(*r.pending)
	if err != nil {
		return 0, nil, err
	}
	version := r.loaded.Version + 1
	if r.loaded.Exists {
		tag, err := r.tx.Exec(ctx, `UPDATE cafe.aggregates SET version=$3,state=$4 WHERE kind=$1 AND id=$2 AND version=$5`, r.kind, r.target, version, data, r.loaded.Version)
		if err != nil {
			return 0, nil, err
		}
		if tag.RowsAffected() != 1 {
			return 0, nil, fmt.Errorf("optimistic update lost its owner version")
		}
	} else if _, err = r.tx.Exec(ctx, `INSERT INTO cafe.aggregates(kind,id,version,state) VALUES($1,$2,$3,$4)`, r.kind, r.target, version, data); err != nil {
		return 0, nil, err
	}
	return version, data, nil
}
