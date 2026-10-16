package messaging

import (
	"context"
	"crypto/rand"
	"fmt"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
	rabbit "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

// pending tracks a request awaiting its correlated transport reply.
type pending struct {
	owner string
	reply chan pb.Reply
}

// RabbitMQRequestClient owns the development API's exclusive reply consumers. Concurrent
// requests are matched by transport identity; command identity stays independent.
type RabbitMQRequestClient struct {
	mu         sync.Mutex
	publishing chan struct{}
	channel    *rabbit.Channel
	returns    <-chan rabbit.Return
	routes     map[string]string
	waiting    map[string]pending
}

func NewClient() *RabbitMQRequestClient {
	return &RabbitMQRequestClient{waiting: map[string]pending{}, publishing: make(chan struct{}, 1)}
}
func (c *RabbitMQRequestClient) Call(ctx context.Context, request pb.Request) (pb.Reply, error) {
	if _, _, err := Validate(request, request.GetContext()); err != nil {
		return nil, err
	}
	kind := "query"
	if pb.Command(request) != nil {
		kind = "command"
	}
	data, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	response := make(chan pb.Reply, 1)
	c.mu.Lock()
	channel, returns, route := c.channel, c.returns, c.routes[request.GetContext()]
	if channel == nil || route == "" {
		c.mu.Unlock()
		return nil, fmt.Errorf("request broker unavailable")
	}
	if _, exists := c.waiting[request.GetRequestId()]; exists {
		c.mu.Unlock()
		return nil, fmt.Errorf("duplicate request identity")
	}
	c.waiting[request.GetRequestId()] = pending{request.GetContext(), response}
	c.mu.Unlock()
	defer func() { c.mu.Lock(); delete(c.waiting, request.GetRequestId()); c.mu.Unlock() }()
	select {
	case c.publishing <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if ctx.Err() != nil {
		<-c.publishing
		return nil, ctx.Err()
	}
	confirmation, err := channel.PublishWithDeferredConfirmWithContext(ctx, "cafe.requests", "request."+request.GetContext()+"."+kind, true, false, rabbit.Publishing{ContentType: "application/x-protobuf", DeliveryMode: rabbit.Persistent, MessageId: request.GetRequestId(), CorrelationId: request.GetRequestId(), AppId: "api", ReplyTo: route, Type: kind, Expiration: "15000", Body: data})
	if err == nil {
		if confirmation == nil {
			err = fmt.Errorf("publisher confirms unavailable")
		} else {
			var accepted bool
			accepted, err = confirmation.WaitContext(ctx)
			if err == nil && !accepted {
				err = fmt.Errorf("request rejected")
			}
			if err == nil {
				select {
				case <-returns:
					err = fmt.Errorf("request unroutable")
				default:
				}
			}
		}
	}
	<-c.publishing
	if err != nil {
		_ = channel.Close()
		return nil, err
	}
	select {
	case reply := <-response:
		if reply == nil {
			return nil, fmt.Errorf("request connection lost")
		}
		return reply, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (c *RabbitMQRequestClient) receive(owner, route string, delivery rabbit.Delivery) {
	reply, err := pb.NewReply(owner, "")
	if err != nil {
		_ = delivery.Ack(false)
		return
	}
	valid := delivery.ContentType == "application/x-protobuf" && delivery.DeliveryMode == rabbit.Persistent && delivery.Type == "reply" && delivery.AppId == owner && delivery.RoutingKey == route && delivery.MessageId == delivery.CorrelationId && proto.Unmarshal(delivery.Body, reply) == nil && reply.GetContractVersion() == 1 && reply.GetContext() == owner && reply.GetRequestId() == delivery.CorrelationId
	if valid {
		c.mu.Lock()
		attempt, exists := c.waiting[reply.GetRequestId()]
		if exists && attempt.owner == owner {
			select {
			case attempt.reply <- reply:
			default:
			}
		}
		c.mu.Unlock()
	}
	// Late or duplicate replies have no pending caller and can be discarded.
	_ = delivery.Ack(false)
}
func (c *RabbitMQRequestClient) Run(ctx context.Context, url string, owners []string) {
	for ctx.Err() == nil {
		_ = c.connected(ctx, url, owners)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
func (c *RabbitMQRequestClient) connected(ctx context.Context, url string, owners []string) error {
	conn, err := rabbit.DialConfig(url, rabbit.Config{Dial: rabbit.DefaultDial(5 * time.Second)})
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = channel.Close() }()
	if err = channel.Confirm(false); err != nil {
		return err
	}
	receiver, err := conn.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = receiver.Close() }()
	if err = receiver.Qos(128, 0, false); err != nil {
		return err
	}
	instance := newConnectionID()
	routes := make(map[string]string, len(owners))
	for _, owner := range owners {
		queue, route := "ref.api."+owner+".replies."+instance, "reply."+owner+"."+instance
		if _, err = receiver.QueueDeclare(queue, false, true, true, false,
			rabbit.Table{"x-queue-type": "classic", "x-message-ttl": int32(30000)}); err != nil {
			return err
		}
		if err = receiver.QueueBind(queue, route, "cafe.replies", false, nil); err != nil {
			return err
		}
		routes[owner] = route
		deliveries, err := receiver.Consume(queue, "", false, true, false, false, nil)
		if err != nil {
			return err
		}
		go func() {
			for delivery := range deliveries {
				c.receive(owner, route, delivery)
			}
		}()
	}
	closed := conn.NotifyClose(make(chan *rabbit.Error, 1))
	channelClosed := channel.NotifyClose(make(chan *rabbit.Error, 1))
	receiverClosed := receiver.NotifyClose(make(chan *rabbit.Error, 1))
	c.mu.Lock()
	c.channel = channel
	c.routes = routes
	c.returns = channel.NotifyReturn(make(chan rabbit.Return, 1))
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.channel = nil
		c.routes = nil
		for _, attempt := range c.waiting {
			select {
			case attempt.reply <- nil:
			default:
			}
		}
		c.mu.Unlock()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-closed:
		if err != nil {
			return err
		}
	case err := <-channelClosed:
		if err != nil {
			return err
		}
	case err := <-receiverClosed:
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("request connection closed")
}

// Reconnects get a fresh address so retired connections cannot consume new replies.
func newConnectionID() string {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		panic(err)
	}
	id[6] = id[6]&15 | 64
	id[8] = id[8]&63 | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[:4], id[4:6], id[6:8], id[8:10], id[10:])
}
