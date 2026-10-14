package postgres

import (
	"context"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"regexp"
	"strings"
)

// PublicationOutbox atomically completes a subscription and stages its outgoing integration event.
type PublicationOutbox[P any] struct {
	db             *ContextDatabase
	consumer, kind string
	publication    func(P) a.Publication
}

// NewPublicationOutbox binds an owner subscription to its typed published contract.
func NewPublicationOutbox[P any](db *ContextDatabase, consumer, kind string, publication func(P) a.Publication) *PublicationOutbox[P] {
	if !strings.HasPrefix(consumer, db.owner+".") {
		panic("foreign publication subscription")
	}
	return &PublicationOutbox[P]{db: db, consumer: consumer, kind: kind, publication: publication}
}
func (p *PublicationOutbox[P]) Publish(ctx context.Context, m a.Metadata, payload P) (a.Outcome, error) {
	for _, id := range []string{m.ID, m.AggregateID, m.SourceEventID, m.CorrelationID} {
		if err := d.ValidateID(id); err != nil {
			return a.Outcome{}, err
		}
	}
	if m.Consumer != p.consumer || m.SourceKind != p.kind || m.SourceVersion == 0 || m.SourceVersion > 9007199254740991 || m.SourceTime.IsZero() || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(m.SourceHash) {
		return a.Outcome{}, fmt.Errorf("invalid publication receipt metadata")
	}
	tx, err := begin(ctx, p.db.pool)
	if err != nil {
		return a.Outcome{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	target := "publication:" + p.kind + ":" + m.AggregateID
	done, outcome, err := loadIncoming(ctx, tx, m, target, m.AggregateID)
	if err != nil {
		return a.Outcome{}, err
	}
	if done {
		return outcome, tx.Commit(ctx)
	}
	publication := p.publication(payload)
	if publication.Visibility != a.Public {
		return a.Outcome{}, fmt.Errorf("publication handler requires an integration event")
	}
	message := a.Message{ID: NewID(), Name: publication.Name, Context: p.db.owner, Visibility: a.Public, ContractVersion: 1, AggregateKind: p.kind, AggregateID: m.AggregateID, AggregateVersion: m.SourceVersion, CorrelationID: m.CorrelationID, CausationID: m.SourceEventID, OccurredAt: m.SourceTime, Payload: publication.Payload}
	if err = p.db.appendEvent(ctx, tx, message); err != nil {
		return a.Outcome{}, err
	}
	outcome = a.Outcome{AggregateID: m.AggregateID, Version: m.SourceVersion, Status: "publication_queued"}
	if err = recordIncoming(ctx, tx, m, target, outcome); err != nil {
		return a.Outcome{}, err
	}
	return outcome, tx.Commit(ctx)
}
