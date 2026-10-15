//go:build integration

package amqp_test

import (
	"bytes"
	"context"
	"fmt"
	transport "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/amqp"
	internal "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/internalcommands"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	ownercodec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	commands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	codec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	rabbit "github.com/rabbitmq/amqp091-go"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPrivateCommandRetryReplayPreservesAcceptedIdentity(t *testing.T) {
	ctx, cancelAll := context.WithTimeout(t.Context(), 25*time.Second)
	defer cancelAll()
	url := os.Getenv("ORDERING_BROKER_URL")
	private := ownercodec.RecordCollectionCodec{}
	consumer := private.Consumer()
	source := store.NewID()
	metadata := a.Metadata{ID: a.DerivedID(consumer, source), Name: consumer, Consumer: consumer, SourceEventID: source, CausationID: source, SourceHash: strings.Repeat("a", 64), AggregateID: store.NewID(), CorrelationID: store.NewID(), Input: map[string]any{"evidence": "original"}}
	body, err := private.Encode(metadata, commands.RecordCollectionCommand{CustomerID: store.NewID()})
	if err != nil {
		t.Fatal(err)
	}
	decoder := internal.DeliveryDecoder[commands.RecordCollectionCommand](private)
	_, message, err := decoder(body)
	if err != nil {
		t.Fatal(err)
	}
	sub := transport.Subscription{Binding: transport.Binding{Consumer: consumer + ".command", Event: consumer + ".command", Visibility: "command", Context: "ordering"}, DecodeCommand: decoder}
	var failing atomic.Bool
	failing.Store(true)
	var calls atomic.Int32
	delivered := make(chan struct{}, 1)
	sub.Handle = func(_ context.Context, received a.Metadata, m a.Message) error {
		calls.Add(1)
		if received.ID != metadata.ID || received.SourceEventID != source || received.SourceHash != metadata.SourceHash || received.AggregateID != metadata.AggregateID {
			t.Error("command consumer rewrote accepted evidence")
		}
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
	go func() { done <- transport.Consume(workerCtx, url, sub, codec.Decode, store.DerivedID) }()
	publisher, err := transport.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	if err = publisher.Publish(ctx, message, body); err != nil {
		t.Fatal(err)
	}
	publisher.Close()
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
		info, err := channel.QueueInspect(transport.Queue(sub.Binding.Consumer) + ".dead")
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
	dead, ok, err := channel.Get(transport.Queue(sub.Binding.Consumer)+".dead", false)
	if err != nil || !ok || !bytes.Equal(dead.Body, body) || dead.MessageId != message.ID {
		t.Fatal("dead letter lost original bytes or identity")
	}
	if err = dead.Nack(false, true); err != nil {
		t.Fatal(err)
	}
	failing.Store(false)
	if replayed, err := transport.Replay(ctx, url, sub.Binding.Consumer); err != nil || !replayed {
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
