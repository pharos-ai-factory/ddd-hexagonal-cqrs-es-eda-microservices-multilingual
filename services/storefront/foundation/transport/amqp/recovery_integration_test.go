//go:build integration

package amqp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	codec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	support "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
	rabbit "github.com/rabbitmq/amqp091-go"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func fixture(t *testing.T) (context.Context, string, Subscription, a.Message, []byte) {
	t.Helper()
	return ownedFixture(t, "menu")
}
func ownedFixture(t *testing.T, owner string) (context.Context, string, Subscription, a.Message, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Second)
	t.Cleanup(cancel)
	url := os.Getenv(strings.ToUpper(owner) + "_BROKER_URL")
	if url == "" {
		t.Fatalf("integration lane requires %s_BROKER_URL", strings.ToUpper(owner))
	}
	consumer := owner + ".test-" + store.NewID()
	binding := Binding{Consumer: consumer, Context: owner, Event: "menu.drink-published", Visibility: "domain"}
	if owner == "ordering" {
		binding.Event, binding.Visibility = "ordering.order-placed", "integration"
	}
	if err := Declare(os.Getenv("BROKER_ADMIN_URL"), []Binding{binding}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn, err := rabbit.Dial(os.Getenv("BROKER_ADMIN_URL"))
		if err != nil {
			return
		}
		defer conn.Close()
		channel, err := conn.Channel()
		if err != nil {
			return
		}
		defer channel.Close()
		for _, suffix := range []string{"", ".retry", ".dead"} {
			_, _ = channel.QueueDelete(Queue(consumer)+suffix, false, false, false)
		}
	})
	account := store.NewID()
	message := a.Message{ID: store.NewID(), Name: binding.Event, Context: "menu", Visibility: a.Private, ContractVersion: 1, AggregateKind: "drink", AggregateID: account, AggregateVersion: 3, CorrelationID: store.NewID(), CausationID: store.NewID(), OccurredAt: time.Now().UTC(), Payload: menu.DrinkPublished{DrinkID: account, Name: "Coffee", Revision: 1}}
	if owner == "ordering" {
		message.Context, message.Visibility, message.AggregateKind = owner, a.Public, "order"
		message.Payload = model.OrderPlaced{OrderID: account, CustomerID: store.NewID(), EditionID: store.NewID(), Currency: "GBP",
			Lines: []model.Line{{ID: store.NewID(), OfferCode: "COFFEE", Name: "Coffee", Quantity: 1, Minor: 300}}}
	}
	body, err := codec.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, url, Subscription{Binding: binding}, message, body
}
func publishDirect(t *testing.T, ctx context.Context, url, key string, m a.Message, body []byte) {
	t.Helper()
	p, err := Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	err = p.send(ctx, DeliveryExchange(m.Context), key, rabbit.Publishing{ContentType: "application/x-protobuf", DeliveryMode: rabbit.Persistent, MessageId: m.ID, Type: m.Name, AppId: m.Context, CorrelationId: m.CorrelationID, Body: body, Headers: rabbit.Table{"contract-version": int32(1)}})
	if err != nil {
		t.Fatal(err)
	}
}
func TestMandatoryPublicationIsNotSuccessfulWhenUnroutable(t *testing.T) {
	ctx, url, _, m, body := fixture(t)
	p, err := Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	err = p.send(ctx, Exchange, "domain.menu.unbound-"+store.NewID(), rabbit.Publishing{ContentType: "application/x-protobuf", DeliveryMode: rabbit.Persistent, MessageId: m.ID, Body: body})
	if err == nil || !strings.Contains(err.Error(), "returned") {
		t.Fatalf("unroutable mandatory publication: %v", err)
	}
}
func TestCommitThenLostAcknowledgementDoesNotRepeatEffect(t *testing.T) {
	ctx, url, sub, message, body := ownedFixture(t, "ordering")
	db, err := store.Open(ctx, os.Getenv("ORDERING_DATABASE_URL"), "ordering", codec.Encode)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	type Count struct{ Value int }
	commands := support.SnapshotDecisionFixture[Count]{Transaction: store.NewAggregateTransaction[Count, Count](db, "test_counter", func(repo a.WriteRepository[Count]) a.WriteRepository[Count] { return repo })}
	queries := store.NewSnapshotReadRepository[Count](db, "test_counter")
	target := store.NewID()
	var decisions atomic.Int32
	apply := func(m a.Metadata) error {
		m.AggregateID = target
		_, err := commands.Execute(ctx, m, func(s a.Loaded[Count]) (a.Mutation[Count], error) {
			decisions.Add(1)
			return a.Changed(Count{s.State.Value + 1}, "active"), nil
		})
		return err
	}
	conn, err := rabbit.Dial(url)
	if err != nil {
		t.Fatal(err)
	}
	channel, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	deliveries, err := channel.Consume(Queue(sub.Binding.Consumer), "", false, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	publishDirect(t, ctx, url, Queue(sub.Binding.Consumer), message, body)
	var delivery rabbit.Delivery
	select {
	case delivery = <-deliveries:
	case <-ctx.Done():
		t.Fatal("first delivery timed out")
	}
	if !bytes.Equal(delivery.Body, body) {
		t.Fatal("wire bytes changed")
	}
	sum := sha256.Sum256(body)
	metadata := a.Metadata{ID: store.DerivedID(sub.Binding.Consumer, message.ID), Name: sub.Binding.Consumer, CorrelationID: message.CorrelationID, CausationID: message.ID, Input: message.Payload, Consumer: sub.Binding.Consumer, SourceEventID: message.ID, SourceHash: hex.EncodeToString(sum[:])}
	if err = apply(metadata); err != nil {
		t.Fatal(err)
	}
	// Close a real connection after the database commit without acknowledging.
	_ = conn.Close()
	received := make(chan error, 1)
	sub.Handle = func(_ context.Context, m a.Metadata, _ a.Message) error { err := apply(m); received <- err; return err }
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Consume(workerCtx, url, sub, codec.Decode, store.DerivedID) }()
	select {
	case err = <-received:
		if err != nil {
			t.Fatal(err)
		}
	case err = <-done:
		t.Fatalf("consumer stopped: %v", err)
	case <-ctx.Done():
		t.Fatal("redelivery timed out")
	}
	loaded, err := queries.Get(ctx, target)
	if err != nil || loaded.State.Value != 1 || loaded.Version != 1 || decisions.Load() != 1 {
		t.Fatalf("duplicate effect: %+v decisions=%d err=%v", loaded, decisions.Load(), err)
	}
	cancel()
	<-done
}
func TestBoundedRetryDeadLetterAndExplicitReplay(t *testing.T) {
	ctx, url, sub, message, body := fixture(t)
	var failing atomic.Bool
	failing.Store(true)
	var calls atomic.Int32
	delivered := make(chan struct{}, 1)
	sub.Handle = func(_ context.Context, _ a.Metadata, m a.Message) error {
		calls.Add(1)
		if m.ID != message.ID {
			t.Error("retry changed message identity")
		}
		if failing.Load() {
			return fmt.Errorf("injected database outage")
		}
		delivered <- struct{}{}
		return nil
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Consume(workerCtx, url, sub, codec.Decode, store.DerivedID) }()
	publishDirect(t, ctx, url, Queue(sub.Binding.Consumer), message, body)
	admin, err := rabbit.Dial(os.Getenv("BROKER_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	channel, err := admin.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	for {
		info, err := channel.QueueInspect(Queue(sub.Binding.Consumer) + ".dead")
		if err != nil {
			t.Fatal(err)
		}
		if info.Messages == 1 {
			break
		}
		select {
		case err = <-done:
			t.Fatalf("consumer exited: %v", err)
		case <-ctx.Done():
			t.Fatal("dead-letter deadline exceeded")
		case <-time.After(50 * time.Millisecond):
		}
	}
	if calls.Load() != 4 {
		t.Fatalf("wanted initial attempt plus three retries, got %d", calls.Load())
	}
	dead, ok, err := channel.Get(Queue(sub.Binding.Consumer)+".dead", false)
	if err != nil || !ok || !bytes.Equal(dead.Body, body) || dead.MessageId != message.ID {
		t.Fatal("dead letter lost original bytes or identity")
	}
	if err = dead.Nack(false, true); err != nil {
		t.Fatal(err)
	}
	failing.Store(false)
	if replayed, err := Replay(ctx, url, sub.Binding.Consumer); err != nil || !replayed {
		t.Fatalf("replay failed: %t %v", replayed, err)
	}
	select {
	case <-delivered:
	case err = <-done:
		t.Fatalf("replay consumer exited: %v", err)
	case <-ctx.Done():
		t.Fatal("replay timed out")
	}
	cancel()
	<-done
}
func TestPrivateDomainBindingIsDeniedToAnotherContext(t *testing.T) {
	ctx, _, _, _, _ := fixture(t)
	_ = ctx
	conn, err := rabbit.Dial(os.Getenv("ORDERING_BROKER_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	queue := "ref.ordering.private-test-" + store.NewID()
	admin, err := rabbit.Dial(os.Getenv("BROKER_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	topology, err := admin.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer topology.Close()
	if _, err = topology.QueueDeclare(queue, true, false, false, false, rabbit.Table{"x-queue-type": "quorum"}); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = topology.QueueDelete(queue, false, false, false) }()
	err = channel.QueueBind(queue, "domain.loyalty.reward-earned", Exchange, false, nil)
	if err == nil {
		t.Fatal("Ordering bound a private Loyalty domain event")
	}
}
