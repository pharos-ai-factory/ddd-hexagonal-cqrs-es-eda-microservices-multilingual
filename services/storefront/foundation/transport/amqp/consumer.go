package amqp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/diagnostics"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	rabbit "github.com/rabbitmq/amqp091-go"
	"strings"
	"time"
)

type Handler func(context.Context, a.Metadata, a.Message) error
type Subscription struct {
	Binding Binding
	Handle  Handler
}

func Consume(ctx context.Context, url string, sub Subscription, decode a.Decoder, ids a.IDFactory) error {
	publisher, err := Connect(url)
	if err != nil {
		return err
	}
	defer publisher.Close()
	channel, err := publisher.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	if err = channel.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := channel.Consume(Queue(sub.Binding.Consumer), "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, open := <-deliveries:
			if !open {
				return fmt.Errorf("consumer connection closed")
			}
			attempt, validation := retryAttempt(delivery.Headers)
			var message a.Message
			if validation == nil {
				message, validation = decode(delivery.Body)
			}
			if validation == nil {
				validation = validateDelivery(delivery, message, sub.Binding)
			}
			processing := validation
			if processing == nil {
				sum := sha256.Sum256(delivery.Body)
				metadata := a.Metadata{ID: ids(sub.Binding.Consumer, message.ID), Name: sub.Binding.Consumer, CorrelationID: message.CorrelationID, CausationID: message.ID, Input: message.Payload, Consumer: sub.Binding.Consumer, SourceEventID: message.ID, SourceHash: hex.EncodeToString(sum[:])}
				processing = sub.Handle(ctx, metadata, message)
			}
			if processing == nil {
				if err = delivery.Ack(false); err != nil {
					return err
				}
				continue
			}
			target := Queue(sub.Binding.Consumer) + ".retry"
			var violation *d.Violation
			if validation != nil || errors.As(processing, &violation) || attempt >= 3 {
				target = Queue(sub.Binding.Consumer) + ".dead"
			}
			headers := rabbit.Table{}
			for key, value := range delivery.Headers {
				if key != "x-death" {
					headers[key] = value
				}
			}
			headers["ref-attempt"] = int32(attempt + 1)
			headers["ref-failure"] = diagnostics.Class(processing)
			outgoing := rabbit.Publishing{ContentType: delivery.ContentType, DeliveryMode: rabbit.Persistent, MessageId: delivery.MessageId, Type: delivery.Type, AppId: delivery.AppId, CorrelationId: delivery.CorrelationId, Body: delivery.Body, Headers: headers}
			publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = publisher.send(publishCtx, DeliveryExchange(sub.Binding.Context), target, outgoing)
			cancel()
			if err != nil {
				_ = delivery.Nack(false, true)
				return err
			}
			diagnostics.Record(sub.Binding.Context, sub.Binding.Consumer, message.ID, message.CorrelationID, processing, strings.HasSuffix(target, ".dead"))
			if err = delivery.Ack(false); err != nil {
				return err
			}
		}
	}
}
func retryAttempt(headers rabbit.Table) (int, error) {
	value, present := headers["ref-attempt"]
	if !present {
		return 0, nil
	}
	attempt, ok := value.(int32)
	if !ok || attempt < 0 || attempt > 3 {
		return 0, fmt.Errorf("invalid retry counter")
	}
	return int(attempt), nil
}
func validateDelivery(delivery rabbit.Delivery, m a.Message, b Binding) error {
	if m.Name != b.Event {
		return fmt.Errorf("unexpected event for consumer")
	}
	if m.Visibility == a.Private && m.Context != b.Context {
		return fmt.Errorf("private domain event crossed its context boundary")
	}
	version, ok := delivery.Headers["contract-version"].(int32)
	if !ok || version != int32(m.ContractVersion) || delivery.ContentType != "application/x-protobuf" || delivery.DeliveryMode != rabbit.Persistent || delivery.MessageId != m.ID || delivery.Type != m.Name || delivery.AppId != m.Context || delivery.CorrelationId != m.CorrelationID {
		return fmt.Errorf("AMQP metadata disagrees with the event envelope")
	}
	return nil
}

// Replay moves one quarantined message back to its original consumer after a
// confirmed publication. Its identity and payload are preserved; attempts reset.
func Replay(ctx context.Context, url, consumer string) (bool, error) {
	publisher, err := Connect(url)
	if err != nil {
		return false, err
	}
	defer publisher.Close()
	delivery, ok, err := publisher.channel.Get(Queue(consumer)+".dead", false)
	if err != nil || !ok {
		return false, err
	}
	headers := rabbit.Table{}
	for key, value := range delivery.Headers {
		if key != "ref-attempt" && key != "ref-failure" && key != "x-death" {
			headers[key] = value
		}
	}
	message := rabbit.Publishing{ContentType: delivery.ContentType, DeliveryMode: rabbit.Persistent, MessageId: delivery.MessageId, Type: delivery.Type, AppId: delivery.AppId, CorrelationId: delivery.CorrelationId, Body: delivery.Body, Headers: headers}
	if err = publisher.send(ctx, DeliveryExchange(strings.SplitN(consumer, ".", 2)[0]), Queue(consumer), message); err != nil {
		_ = delivery.Nack(false, true)
		return false, err
	}
	return true, delivery.Ack(false)
}
