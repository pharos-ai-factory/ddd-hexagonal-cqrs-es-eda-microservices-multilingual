package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

// Dispatch carries immutable event bytes and a fenced outbox lease.
type Dispatch struct {
	EventID    string
	Name       string
	Body       []byte
	Token      string
	Generation int64
}

func (db *ContextDatabase) Claim(ctx context.Context, lease time.Duration) (Dispatch, bool, error) {
	token := NewID()
	var d Dispatch
	err := db.pool.QueryRow(ctx, `WITH candidate AS (
 SELECT event_id FROM cafe.dispatches WHERE completed_at IS NULL AND available_at<=clock_timestamp()
 AND (lease_until IS NULL OR lease_until<clock_timestamp()) ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
 ), claimed AS (
 UPDATE cafe.dispatches d SET lease_token=$1,lease_until=clock_timestamp()+$2::interval,generation=generation+1,attempts=attempts+1
 FROM candidate c WHERE d.event_id=c.event_id RETURNING d.event_id,d.generation
 ) SELECT o.id,o.event_name,o.body,c.generation FROM claimed c JOIN cafe.outbox_events o ON o.id=c.event_id`, token, lease.String()).Scan(&d.EventID, &d.Name, &d.Body, &d.Generation)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, false, nil
	}
	d.Token = token
	return d, err == nil, err
}
func (db *ContextDatabase) Complete(ctx context.Context, d Dispatch) (bool, error) {
	tag, err := db.pool.Exec(ctx, `UPDATE cafe.dispatches SET completed_at=clock_timestamp(),lease_until=NULL,lease_token=NULL,last_error=NULL WHERE event_id=$1 AND lease_token=$2 AND generation=$3 AND lease_until>clock_timestamp()`, d.EventID, d.Token, d.Generation)
	return err == nil && tag.RowsAffected() == 1, err
}
func (db *ContextDatabase) Retry(ctx context.Context, d Dispatch, reason string) (bool, error) {
	tag, err := db.pool.Exec(ctx, `UPDATE cafe.dispatches SET available_at=clock_timestamp()+interval '1 second',lease_until=NULL,lease_token=NULL,last_error=$4 WHERE event_id=$1 AND lease_token=$2 AND generation=$3 AND lease_until>clock_timestamp()`, d.EventID, d.Token, d.Generation, reason)
	return err == nil && tag.RowsAffected() == 1, err
}
func (db *ContextDatabase) Pending(ctx context.Context) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, `SELECT count(*) FROM cafe.dispatches WHERE completed_at IS NULL`).Scan(&count)
	return count, err
}
