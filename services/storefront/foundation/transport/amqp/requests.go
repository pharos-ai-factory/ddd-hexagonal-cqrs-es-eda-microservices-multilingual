package amqp

import (
	"context"
	"fmt"
	rabbit "github.com/rabbitmq/amqp091-go"
	"time"
)

const RequestsExchange = "cafe.requests"
const RepliesExchange = "cafe.replies"

func RequestQueue(owner string, kinds ...string) string {
	kind := "command"
	if len(kinds) > 0 {
		kind = kinds[0]
	}
	suffix := "commands"
	if kind == "query" {
		suffix = "queries"
	}
	return "ref." + owner + "." + suffix
}
func ReplyQueue(owner string) string { return "ref.api." + owner + ".replies" }

// DeclareRequests is bootstrap-only. Runtime credentials cannot change topology.
func DeclareRequests(url string, owners []string) error {
	conn, err := rabbit.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	for _, name := range []string{RequestsExchange, RepliesExchange} {
		if err = channel.ExchangeDeclare(name, "topic", true, false, false, false, nil); err != nil {
			return err
		}
	}
	for _, owner := range owners {
		for _, kind := range []string{"command", "query"} {
			queue := RequestQueue(owner, kind)
			exchange := DeliveryExchange(owner)
			if err = channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
				return err
			}
			deadArgs := rabbit.Table{"x-queue-type": "quorum", "x-delivery-limit": int32(-1)}
			if _, err = channel.QueueDeclare(queue+".dead", true, false, false, false, deadArgs); err != nil {
				return err
			}
			if err = channel.QueueBind(queue+".dead", queue+".dead", exchange, false, nil); err != nil {
				return err
			}
			args := rabbit.Table{"x-queue-type": "quorum", "x-delivery-limit": int32(5), "x-message-ttl": int32(15000), "x-dead-letter-exchange": exchange, "x-dead-letter-routing-key": queue + ".dead", "x-dead-letter-strategy": "at-least-once", "x-overflow": "reject-publish"}
			if _, err = channel.QueueDeclare(queue, true, false, false, false, args); err != nil {
				return err
			}
			if err = channel.QueueBind(queue, "request."+owner+"."+kind, RequestsExchange, false, nil); err != nil {
				return err
			}

		}
		if _, err = channel.QueueDeclare(ReplyQueue(owner), true, false, false, false, rabbit.Table{"x-queue-type": "quorum", "x-message-ttl": int32(30000), "x-delivery-limit": int32(5)}); err != nil {
			return err
		}
		if err = channel.QueueBind(ReplyQueue(owner), "reply."+owner, RepliesExchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}

type RequestHandler func(context.Context, string, []byte) ([]byte, error)

// ServeRequests has one unacknowledged request per owner and kind. Commands
// acknowledge committed reply intent; queries acknowledge confirmed publication.
func ServeRequests(ctx context.Context, url, owner, kind string, handle RequestHandler) error {
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
	deliveries, err := channel.Consume(RequestQueue(owner, kind), "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, open := <-deliveries:
			if !open {
				return fmt.Errorf("request consumer closed")
			}
			var body []byte
			if delivery.ContentType != "application/x-protobuf" || delivery.DeliveryMode != rabbit.Persistent || delivery.Type != kind || delivery.AppId != "api" || delivery.MessageId == "" || delivery.CorrelationId != delivery.MessageId || delivery.RoutingKey != "request."+owner+"."+kind || len(delivery.Body) > 65536 {
				err = fmt.Errorf("invalid request delivery")
			} else {
				body, err = handle(ctx, delivery.MessageId, delivery.Body)
			}
			publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if err != nil {
				outgoing := rabbit.Publishing{ContentType: delivery.ContentType, DeliveryMode: rabbit.Persistent, MessageId: delivery.MessageId, Type: delivery.Type, AppId: delivery.AppId, CorrelationId: delivery.CorrelationId, Body: delivery.Body, Headers: delivery.Headers}
				err = publisher.send(publishCtx, DeliveryExchange(owner), RequestQueue(owner, kind)+".dead", outgoing)
			} else if len(body) > 0 {
				err = publisher.send(publishCtx, RepliesExchange, "reply."+owner, rabbit.Publishing{ContentType: "application/x-protobuf", DeliveryMode: rabbit.Persistent, MessageId: delivery.MessageId, CorrelationId: delivery.MessageId, Type: "reply", AppId: owner, Body: body})
			}
			cancel()
			if err != nil {
				_ = delivery.Nack(false, true)
				return err
			}
			if err = delivery.Ack(false); err != nil {
				return err
			}
		}
	}
}

func (p *ConfirmedPublisher) PublishReply(ctx context.Context, owner, id string, body []byte) error {
	return p.send(ctx, RepliesExchange, "reply."+owner, rabbit.Publishing{ContentType: "application/x-protobuf", DeliveryMode: rabbit.Persistent, MessageId: id, CorrelationId: id, Type: "reply", AppId: owner, Body: body})
}
