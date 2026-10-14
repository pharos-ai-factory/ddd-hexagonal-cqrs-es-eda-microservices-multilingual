//go:build integration

package amqp

import (
	"context"

	rabbit "github.com/rabbitmq/amqp091-go"
)

// SendForTest exercises mandatory confirms with deliberately invalid metadata.
func SendForTest(publisher *ConfirmedPublisher, ctx context.Context, exchange, key string, message rabbit.Publishing) error {
	return publisher.send(ctx, exchange, key, message)
}

// GetForTest observes quarantine without exposing the channel in production.
func GetForTest(publisher *ConfirmedPublisher, queue string, autoAck bool) (rabbit.Delivery, bool, error) {
	return publisher.channel.Get(queue, autoAck)
}
