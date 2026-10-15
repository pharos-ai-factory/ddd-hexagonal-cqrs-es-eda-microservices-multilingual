package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	internal "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/internalcommands"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// CommandOutbox commits receiving evidence, immutable command bytes and dispatch intent together.
type CommandOutbox[C any] struct {
	db    *ContextDatabase
	codec internal.Codec[C]
}

func NewCommandOutbox[C any](db *ContextDatabase, codec internal.Codec[C]) *CommandOutbox[C] {
	if db.owner != codec.Owner() {
		panic("command outbox crossed its owner")
	}
	return &CommandOutbox[C]{db: db, codec: codec}
}
func (p *CommandOutbox[C]) Enqueue(ctx context.Context, m a.Metadata, command C) (a.Outcome, error) {
	if err := internal.Validate(p.db.owner, p.codec.Consumer(), m); err != nil {
		return a.Outcome{}, err
	}
	body, err := p.codec.Encode(m, command)
	if err != nil {
		return a.Outcome{}, err
	}
	tx, err := begin(ctx, p.db.pool)
	if err != nil {
		return a.Outcome{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "handoff:"+m.Consumer+":"+m.SourceEventID); err != nil {
		return a.Outcome{}, err
	}
	var id, hash, target string
	err = tx.QueryRow(ctx, `SELECT id,fingerprint,target FROM cafe.internal_commands WHERE consumer=$1 AND event_id=$2`, m.Consumer, m.SourceEventID).Scan(&id, &hash, &target)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		if _, err = tx.Exec(ctx, `INSERT INTO cafe.internal_commands(id,consumer,event_id,fingerprint,target,body) VALUES($1,$2,$3,$4,$5,$6)`, m.ID, m.Consumer, m.SourceEventID, m.SourceHash, m.AggregateID, body); err != nil {
			return a.Outcome{}, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO cafe.internal_command_dispatches(event_id) VALUES($1)`, m.ID); err != nil {
			return a.Outcome{}, err
		}
	case err != nil:
		return a.Outcome{}, err
	case id != m.ID || hash != m.SourceHash || target != m.AggregateID:
		return a.Outcome{}, fmt.Errorf("conflicting event hand-off identity")
	}
	return a.Outcome{AggregateID: m.AggregateID, Status: "queued"}, tx.Commit(ctx)
}
