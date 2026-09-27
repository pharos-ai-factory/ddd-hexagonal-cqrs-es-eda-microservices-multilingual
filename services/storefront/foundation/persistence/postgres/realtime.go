package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

func (db *Database) appendRealtime(ctx context.Context, tx pgx.Tx, kind, aggregateID string, revision uint64, state []byte) error {
	id := NewID()
	body, err := db.realtime(id, db.owner, kind, aggregateID, revision, state)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO cafe.realtime_publications(id,channel,aggregate_kind,aggregate_id,revision,body)
		VALUES($1,$2,$3,$4,$5,$6)`, id, "cafe:"+db.owner, kind, aggregateID, revision, body)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO cafe.realtime_dispatches(event_id) VALUES($1)`, id)
	return err
}

type RealtimeDispatch struct {
	ID, Channel, Token string
	Body               []byte
	Generation         int64
}

func (db *Database) ClaimRealtime(ctx context.Context) (RealtimeDispatch, bool, error) {
	d := RealtimeDispatch{Token: NewID()}
	err := db.pool.QueryRow(ctx, `WITH candidate AS (
		SELECT event_id FROM cafe.realtime_dispatches WHERE completed_at IS NULL AND available_at<=clock_timestamp()
		AND (lease_until IS NULL OR lease_until<clock_timestamp())
		ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
	), claimed AS (
		UPDATE cafe.realtime_dispatches d SET lease_token=$1,lease_until=clock_timestamp()+interval '30 seconds',
		generation=generation+1 FROM candidate c WHERE d.event_id=c.event_id RETURNING d.event_id,d.generation
	) SELECT o.id,o.channel,o.body,c.generation FROM claimed c JOIN cafe.realtime_publications o ON o.id=c.event_id`,
		d.Token).Scan(&d.ID, &d.Channel, &d.Body, &d.Generation)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, false, nil
	}
	return d, err == nil, err
}
func (db *Database) FinishRealtime(ctx context.Context, d RealtimeDispatch, failure error) error {
	var reason any
	if failure != nil {
		reason = failure.Error()
	}
	_, err := db.pool.Exec(ctx, `UPDATE cafe.realtime_dispatches SET lease_token=NULL,lease_until=NULL,last_error=$4,
		available_at=clock_timestamp()+interval '1 second',
		completed_at=CASE WHEN $4::text IS NULL THEN clock_timestamp() ELSE NULL END
		WHERE event_id=$1 AND lease_token=$2 AND generation=$3 AND lease_until>clock_timestamp()`,
		d.ID, d.Token, d.Generation, reason)
	return err
}
