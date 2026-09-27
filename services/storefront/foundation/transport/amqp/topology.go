package amqp

import (
	rabbit "github.com/rabbitmq/amqp091-go"
)

type Binding struct {
	Consumer   string
	Event      string
	Visibility string
	Context    string
}

func Queue(consumer string) string         { return "ref." + consumer }
func DeliveryExchange(owner string) string { return "ref." + owner + ".delivery" }
func Declare(url string, bindings []Binding) error {
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
	if err = channel.ExchangeDeclare(Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	for _, b := range bindings {
		queue := Queue(b.Consumer)
		exchange := DeliveryExchange(b.Context)
		if err = channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
			return err
		}
		if _, err = channel.QueueDeclare(queue, true, false, false, false, rabbit.Table{"x-queue-type": "quorum", "x-delivery-limit": int32(-1)}); err != nil {
			return err
		}
		if _, err = channel.QueueDeclare(queue+".dead", true, false, false, false, rabbit.Table{"x-queue-type": "quorum", "x-delivery-limit": int32(-1)}); err != nil {
			return err
		}
		args := rabbit.Table{"x-queue-type": "quorum", "x-delivery-limit": int32(-1), "x-message-ttl": int32(1000), "x-dead-letter-exchange": exchange, "x-dead-letter-routing-key": queue, "x-dead-letter-strategy": "at-least-once", "x-overflow": "reject-publish"}
		if _, err = channel.QueueDeclare(queue+".retry", true, false, false, false, args); err != nil {
			return err
		}
		for _, suffix := range []string{"", ".retry", ".dead"} {
			if err = channel.QueueBind(queue+suffix, queue+suffix, exchange, false, nil); err != nil {
				return err
			}
		}
		if err = channel.QueueBind(queue, b.Visibility+"."+b.Event, Exchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}
