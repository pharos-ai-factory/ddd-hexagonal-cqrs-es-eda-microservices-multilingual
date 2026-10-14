//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	h "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/eventhandlers"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	codec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

func TestPublicationReceiptAndOutboxCommitTogether(t *testing.T) {
	db := fixtureDB(t)
	db.encode = codec.Encode
	owner := admin(t)
	id, sourceID := NewID(), NewID()
	clean(t, owner, id)
	fact := d.OrderPlacedDomainEvent{State: d.OrderState{ID: id, CustomerID: NewID(), EditionID: NewID(), Currency: "EUR", Status: "placed", Lines: []d.LineState{{ID: NewID(), Selection: d.Selection{OfferCode: "TEA", Name: "Tea", Minor: 200}, Quantity: 1}}}}
	// Technical snapshot fixture supplies the source aggregate referenced by the outbox.
	if _, err := snapshotDecisions[d.OrderState](db, "order").Execute(t.Context(), metadata(id, 0), func(a.Loaded[d.OrderState]) (a.Mutation[d.OrderState], error) {
		return a.Changed(fact.State, "placed"), nil
	}); err != nil {
		t.Fatal(err)
	}
	original := a.Message{ID: sourceID, Name: "ordering.order-placed.domain", Context: "ordering", Visibility: a.Private, ContractVersion: 1, AggregateKind: "order", AggregateID: id, AggregateVersion: 1, CorrelationID: NewID(), CausationID: NewID(), OccurredAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), Payload: fact}
	body, err := codec.Encode(original)
	if err != nil {
		t.Fatal(err)
	}
	source, err := codec.Decode(body)
	if err != nil {
		t.Fatal(err)
	}
	consumer := "ordering.publish-order-placed"
	m := a.Metadata{ID: NewID(), AggregateID: id, Name: consumer, Consumer: consumer, CorrelationID: source.CorrelationID, SourceEventID: sourceID, SourceHash: fmt.Sprintf("%x", sha256.Sum256(body)), SourceKind: "order", SourceVersion: source.AggregateVersion, SourceTime: source.OccurredAt}
	t.Cleanup(func() {
		_, _ = owner.Exec(context.Background(), `DELETE FROM cafe.consumer_receipts WHERE event_id=$1`, sourceID)
	})
	handler := h.OrderPlacedDomainEventHandler{Publications: NewPublicationOutbox(db, consumer, "order", codec.PublishOrderPlaced)}
	statement := fmt.Sprintf(`ALTER TABLE cafe.consumer_receipts ADD CONSTRAINT reject_publication_probe CHECK(event_id <> '%s') NOT VALID`, sourceID)
	if _, err = owner.Exec(t.Context(), statement); err != nil {
		t.Fatal(err)
	}
	_, err = handler.Handle(t.Context(), m, fact)
	if _, drop := owner.Exec(t.Context(), `ALTER TABLE cafe.consumer_receipts DROP CONSTRAINT reject_publication_probe`); drop != nil {
		t.Fatal(drop)
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.ConstraintName != "reject_publication_probe" {
		t.Fatalf("expected final receipt failure: %v", err)
	}
	count := func(table, field, value string) int {
		t.Helper()
		var n int
		if err := owner.QueryRow(t.Context(), "SELECT count(*) FROM cafe."+table+" WHERE "+field+"=$1", value).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if count("outbox_events", "aggregate_id", id) != 0 || count("consumer_receipts", "event_id", sourceID) != 0 {
		t.Fatal("records escaped rollback")
	}
	var group sync.WaitGroup
	outcomes := make(chan a.Outcome, 2)
	failures := make(chan error, 2)
	for range 2 {
		group.Go(func() { o, e := handler.Handle(t.Context(), m, fact); outcomes <- o; failures <- e })
	}
	group.Wait()
	if e := <-failures; e != nil {
		t.Fatal(e)
	}
	if e := <-failures; e != nil {
		t.Fatal(e)
	}
	first, second := <-outcomes, <-outcomes
	if first != second || count("outbox_events", "aggregate_id", id) != 1 {
		t.Fatal("concurrent delivery duplicated publication")
	}
	var eventID string
	var saved []byte
	if err := owner.QueryRow(t.Context(), `SELECT id,body FROM cafe.outbox_events WHERE aggregate_id=$1`, id).Scan(&eventID, &saved); err != nil {
		t.Fatal(err)
	}
	outgoing, err := codec.Decode(saved)
	if err != nil {
		t.Fatal(err)
	}
	if outgoing.AggregateVersion != 1 || outgoing.CausationID != sourceID || outgoing.CorrelationID != source.CorrelationID || !outgoing.OccurredAt.Equal(source.OccurredAt) || count("dispatches", "event_id", eventID) != 1 {
		t.Fatal("source evidence or dispatch missing")
	}
	m.SourceHash = strings.Repeat("b", 64)
	if _, err = handler.Handle(t.Context(), m, fact); err == nil {
		t.Fatal("conflicting delivery accepted")
	}
	m.SourceHash = fmt.Sprintf("%x", sha256.Sum256(body))
	if outcome, err := handler.Handle(t.Context(), m, fact); err != nil || outcome != first {
		t.Fatalf("lost ACK replay: %v", err)
	}
	var replay []byte
	if err := owner.QueryRow(t.Context(), `SELECT body FROM cafe.outbox_events WHERE id=$1`, eventID).Scan(&replay); err != nil || !bytes.Equal(saved, replay) {
		t.Fatal("replay changed outgoing bytes")
	}
	audit := "ordering.audit-placed"
	m.Consumer, m.ID = audit, NewID()
	other := NewPublicationOutbox[model.OrderPlacedIntegrationEvent](db, audit, "order", codec.PublishOrderPlaced)
	if _, err = other.Publish(t.Context(), m, outgoing.Payload.(model.OrderPlacedIntegrationEvent)); err != nil {
		t.Fatal(err)
	}
	if count("consumer_receipts", "event_id", sourceID) != 2 || count("outbox_events", "aggregate_id", id) != 2 {
		t.Fatal("independent subscription was suppressed")
	}
}
