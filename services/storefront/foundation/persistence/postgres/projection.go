package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type Projection[S any] struct {
	db   *Database
	name string
}

func Project[S any](db *Database, name string) *Projection[S] {
	return &Projection[S]{db: db, name: name}
}
func (p *Projection[S]) Find(ctx context.Context, key string) (S, bool, error) {
	var result S
	var data []byte
	err := p.db.pool.QueryRow(ctx, `SELECT state FROM cafe.projections WHERE name=$1 AND key=$2`, p.name, key).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, false, nil
	}
	if err != nil {
		return result, false, err
	}
	err = json.Unmarshal(data, &result)
	return result, err == nil, err
}
func (p *Projection[S]) Record(ctx context.Context, m a.Metadata, key string, revision uint64, state S) error {
	tx, err := begin(ctx, p.db.pool)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, p.name+":"+key); err != nil {
		return err
	}
	target := "projection:" + p.name + ":" + key
	done, _, err := loadIncoming(ctx, tx, m, target)
	if err != nil {
		return err
	}
	if done {
		return tx.Commit(ctx)
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	var old []byte
	var previous uint64
	err = tx.QueryRow(ctx, `SELECT revision,state FROM cafe.projections WHERE name=$1 AND key=$2`, p.name, key).Scan(&previous, &old)
	if err == nil {
		if revision >= previous {
			var existing S
			if err = json.Unmarshal(old, &existing); err != nil {
				return err
			}
			left, _ := fingerprint(existing)
			right, _ := fingerprint(state)
			if left != right {
				return fmt.Errorf("immutable projection revision changed its meaning")
			}
		}
		if revision > previous {
			if _, err = tx.Exec(ctx, `UPDATE cafe.projections SET revision=$3,state=$4 WHERE name=$1 AND key=$2`, p.name, key, revision, data); err != nil {
				return err
			}
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		if _, err = tx.Exec(ctx, `INSERT INTO cafe.projections(name,key,revision,state) VALUES($1,$2,$3,$4)`, p.name, key, revision, data); err != nil {
			return err
		}
	} else {
		return err
	}
	if err = recordIncoming(ctx, tx, m, target, a.Outcome{AggregateID: key, Version: revision, Status: "projected"}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
