package postgres

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// ReplyIntent is infrastructure-local. Applications remain unaware of RPC transport.
type ReplyIntent struct {
	ID        string
	Route     string
	Encode    func(a.Outcome) ([]byte, error)
	Committed bool
}

// replyKey scopes reply persistence metadata to the receiving transaction context.
type replyKey struct{}

func WithReply(ctx context.Context, reply *ReplyIntent) context.Context {
	return context.WithValue(ctx, replyKey{}, reply)
}
func commitOutcome(ctx context.Context, tx pgx.Tx, outcome a.Outcome) (a.Outcome, error) {
	intent, _ := ctx.Value(replyKey{}).(*ReplyIntent)
	if intent != nil {
		body, err := intent.Encode(outcome)
		if err != nil {
			return a.Outcome{}, err
		}
		if len(body) == 0 {
			return a.Outcome{}, fmt.Errorf("empty command reply")
		}
		if _, err = tx.Exec(ctx, `INSERT INTO cafe.command_replies(id,body,reply_route) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, intent.ID, body, intent.Route); err != nil {
			return a.Outcome{}, err
		}
		var saved []byte
		var route string
		if err = tx.QueryRow(ctx, `SELECT body,reply_route FROM cafe.command_replies WHERE id=$1`, intent.ID).Scan(&saved, &route); err != nil {
			return a.Outcome{}, err
		}
		if !bytes.Equal(body, saved) || route != intent.Route {
			return a.Outcome{}, fmt.Errorf("request identity reused with different reply bytes or route")
		}
		if _, err = tx.Exec(ctx, `INSERT INTO cafe.command_reply_dispatches(event_id) VALUES($1) ON CONFLICT DO NOTHING`, intent.ID); err != nil {
			return a.Outcome{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return a.Outcome{}, err
	}
	if intent != nil {
		intent.Committed = true
	}
	return outcome, nil
}
