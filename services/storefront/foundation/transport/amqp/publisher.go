package amqp

import (
	"context"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	rabbit "github.com/rabbitmq/amqp091-go"
	"sync"
)

const Exchange = "cafe.events"

type Publisher struct {
	conn    *rabbit.Connection
	channel *rabbit.Channel
	returns <-chan rabbit.Return
	lock    sync.Mutex
}

func Connect(url string) (*Publisher, error) {
	conn, err := rabbit.Dial(url)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err = channel.Confirm(false); err != nil {
		conn.Close()
		return nil, err
	}
	return &Publisher{conn: conn, channel: channel, returns: channel.NotifyReturn(make(chan rabbit.Return, 1))}, nil
}
func (p *Publisher) Close() { _ = p.conn.Close() }
func (p *Publisher) Publish(ctx context.Context, m a.Message, body []byte) error {
	return p.send(ctx, Exchange, string(m.Visibility)+"."+m.Name, rabbit.Publishing{ContentType: "application/x-protobuf", DeliveryMode: rabbit.Persistent, MessageId: m.ID, Type: m.Name, AppId: m.Context, CorrelationId: m.CorrelationID, Body: body, Headers: rabbit.Table{"contract-version": int32(m.ContractVersion)}})
}
func (p *Publisher) send(ctx context.Context, exchange, key string, message rabbit.Publishing) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	confirmation, err := p.channel.PublishWithDeferredConfirmWithContext(ctx, exchange, key, true, false, message)
	if err != nil {
		return err
	}
	if confirmation == nil {
		return fmt.Errorf("publisher confirms are unavailable")
	}
	acknowledged, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !acknowledged {
		return fmt.Errorf("broker rejected publication")
	}
	// RabbitMQ sends basic.return before the confirm for an unroutable mandatory publication.
	select {
	case returned := <-p.returns:
		return fmt.Errorf("mandatory publication returned: %d %s", returned.ReplyCode, returned.ReplyText)
	default:
		return nil
	}
}
