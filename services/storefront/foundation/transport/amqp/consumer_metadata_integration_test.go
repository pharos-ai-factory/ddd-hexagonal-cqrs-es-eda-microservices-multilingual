//go:build integration

package amqp

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	codec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	rabbit "github.com/rabbitmq/amqp091-go"
)

func TestNegativeRetryMetadataCannotExtendTheDeliveryBudget(t *testing.T) {
	for name, counter := range map[string]any{
		"negative": int32(-2147483648), "overflow": int32(2147483647),
		"text": "invalid", "fraction": float64(1.5), "boolean": true, "null": nil,
	} {
		t.Run(name, func(t *testing.T) { malformedRetryMetadata(t, counter) })
	}
}

func malformedRetryMetadata(t *testing.T, counter any) {
	t.Helper()
	ctx, url, sub, message, body := fixture(t)
	publisher, err := Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer publisher.Close()
	var calls atomic.Int32
	sub.Handle = func(context.Context, a.Metadata, a.Message) error {
		calls.Add(1)
		return fmt.Errorf("injected infrastructure failure")
	}
	workerCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		_ = Consume(workerCtx, url, sub, codec.Decode, store.DerivedID)
	}()
	defer func() {
		cancel()
		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			t.Error("consumer did not stop")
		}
	}()
	err = publisher.send(ctx, DeliveryExchange(sub.Binding.Context), Queue(sub.Binding.Consumer), rabbit.Publishing{
		ContentType: "application/x-protobuf", DeliveryMode: rabbit.Persistent, MessageId: message.ID,
		Type: message.Name, AppId: message.Context, CorrelationId: message.CorrelationID, Body: body,
		Headers: rabbit.Table{"contract-version": int32(1), "ref-attempt": counter},
	})
	if err != nil {
		t.Fatal(err)
	}
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		dead, found, err := publisher.channel.Get(Queue(sub.Binding.Consumer)+".dead", false)
		if err != nil {
			t.Fatal(err)
		}
		if found {
			if !bytes.Equal(dead.Body, body) || dead.MessageId != message.ID {
				t.Fatal("quarantine changed the original publication")
			}
			if err = dead.Ack(false); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 0 {
				t.Fatalf("malformed metadata reached the handler %d times", calls.Load())
			}
			return
		}
		if calls.Load() > 4 {
			t.Fatalf("malformed metadata extended the four-attempt budget to %d calls", calls.Load())
		}
		select {
		case <-workerCtx.Done():
			t.Fatal("malformed retry metadata never reached quarantine")
		case <-ticker.C:
		}
	}
}
