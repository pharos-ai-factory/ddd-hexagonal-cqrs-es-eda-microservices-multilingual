//go:build integration

package amqp

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	rabbit "github.com/rabbitmq/amqp091-go"
)

func TestRuntimeBrokerCredentialsDeliverWithoutTopologyAdministration(t *testing.T) {
	if !strings.HasPrefix(os.Getenv("CAFE_DISPOSABLE_PROJECT"), "cafe-reference-test-") {
		t.Fatal("permission probes require disposable infrastructure")
	}
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
	for _, owner := range []string{"menu", "ordering", "preparation", "collection", "loyalty", "communication"} {
		t.Run(owner, func(t *testing.T) {
			url := os.Getenv(strings.ToUpper(owner) + "_BROKER_URL")
			queue := "ref." + owner + ".permissions-" + store.NewID()
			key := "integration." + owner + ".permission-probe"
			if _, err := topology.QueueDeclare(queue, false, false, false, false, nil); err != nil {
				t.Fatal(err)
			}
			defer func() { _, _ = topology.QueueDelete(queue, false, false, false) }()
			if err := topology.QueueBind(queue, key, Exchange, false, nil); err != nil {
				t.Fatal(err)
			}
			publisher, err := Connect(url)
			if err != nil {
				t.Fatal(err)
			}
			defer publisher.Close()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			if err := publisher.send(ctx, Exchange, key, rabbit.Publishing{Body: []byte("permission-probe")}); err != nil {
				t.Fatal(err)
			}
			conn, err := rabbit.Dial(url)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			reader, err := conn.Channel()
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			delivery, ok, err := reader.Get(queue, false)
			if err != nil || !ok || string(delivery.Body) != "permission-probe" {
				t.Fatalf("authorised delivery unavailable: %t %v", ok, err)
			}
			if err := delivery.Ack(false); err != nil {
				t.Fatal(err)
			}
			for _, probe := range []struct {
				name string
				run  func(*rabbit.Channel) error
			}{
				{"declare shared exchange", func(c *rabbit.Channel) error {
					return c.ExchangeDeclare(Exchange, "topic", true, false, false, false, nil)
				}},
				{"delete shared exchange", func(c *rabbit.Channel) error { return c.ExchangeDelete(Exchange, true, false) }},
				{"delete own queue", func(c *rabbit.Channel) error { _, err := c.QueueDelete(queue, false, false, false); return err }},
				{"declare own queue", func(c *rabbit.Channel) error {
					_, err := c.QueueDeclare(queue, false, false, false, false, nil)
					return err
				}},
				{"publish foreign context", func(c *rabbit.Channel) error {
					closed := c.NotifyClose(make(chan *rabbit.Error, 1))
					err := c.PublishWithContext(ctx, Exchange, "domain.foreign.event", false, false, rabbit.Publishing{Body: []byte("forbidden")})
					if err != nil {
						return err
					}
					select {
					case denied := <-closed:
						return denied
					case <-ctx.Done():
						return ctx.Err()
					}
				}},
			} {
				t.Run(probe.name, func(t *testing.T) {
					channel, err := conn.Channel()
					if err != nil {
						t.Fatal(err)
					}
					defer channel.Close()
					err = probe.run(channel)
					var denied *rabbit.Error
					if !errors.As(err, &denied) || denied == nil || denied.Code != 403 {
						t.Fatalf("expected ACCESS_REFUSED, got %v", err)
					}
				})
			}
		})
	}
}
