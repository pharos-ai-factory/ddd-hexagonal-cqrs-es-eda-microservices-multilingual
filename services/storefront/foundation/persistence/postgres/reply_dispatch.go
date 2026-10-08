package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

type ReplyDispatch struct {
	Dispatch
	Expired bool
}

func (db *Database) ClaimReply(ctx context.Context) (ReplyDispatch, bool, error) {
	token := NewID()
	var row ReplyDispatch
	err := db.pool.QueryRow(ctx, `WITH candidate AS (
 SELECT event_id FROM cafe.command_reply_dispatches WHERE completed_at IS NULL AND available_at<=clock_timestamp()
 AND (lease_until IS NULL OR lease_until<clock_timestamp()) ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
 ), claimed AS (
 UPDATE cafe.command_reply_dispatches d SET lease_token=$1,lease_until=clock_timestamp()+interval '8 seconds',generation=generation+1,attempts=attempts+1
 FROM candidate c WHERE d.event_id=c.event_id RETURNING d.event_id,d.generation
 ) SELECT o.id,o.body,c.generation,o.expires_at<=clock_timestamp() FROM claimed c JOIN cafe.command_replies o ON o.id=c.event_id`, token).Scan(&row.EventID, &row.Body, &row.Generation, &row.Expired)
	row.Token = token
	if errors.Is(err, pgx.ErrNoRows) {
		return row, false, nil
	}
	return row, err == nil, err
}
func (db *Database) FinishReply(ctx context.Context, row ReplyDispatch, reason string) error {
	_, err := db.pool.Exec(ctx, `UPDATE cafe.command_reply_dispatches SET lease_token=NULL,lease_until=NULL,
 available_at=clock_timestamp()+interval '1 second',last_error=NULLIF($4,''),
 completed_at=CASE WHEN $4='' OR $4='expired' THEN clock_timestamp() ELSE NULL END
 WHERE event_id=$1 AND lease_token=$2 AND generation=$3 AND lease_until>clock_timestamp()`, row.EventID, row.Token, row.Generation, reason)
	return err
}

// Reply lifetime bounds retries. An explicit retry uses a fresh request ID and the stored command outcome.
const ReplyPollInterval = 100 * time.Millisecond
