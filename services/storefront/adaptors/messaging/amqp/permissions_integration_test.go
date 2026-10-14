//go:build integration

package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	rabbit "github.com/rabbitmq/amqp091-go"
)

func permissionConnection(t *testing.T, variable string) *rabbit.Connection {
	t.Helper()
	if !strings.HasPrefix(os.Getenv("CAFE_DISPOSABLE_PROJECT"), "cafe-reference-test-") {
		t.Fatal("permission probes require disposable infrastructure with runtime workers stopped")
	}
	conn, err := rabbit.Dial(os.Getenv(variable))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func permissionChannel(t *testing.T, conn *rabbit.Connection) *rabbit.Channel {
	t.Helper()
	channel, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = channel.Close() })
	return channel
}

func permissionQueues(t *testing.T) map[string][]string {
	t.Helper()
	data, err := os.ReadFile("../../../apps/topology/permissions.generated.json")
	if err != nil {
		t.Fatal(err)
	}
	var policy struct{ Queues map[string][]string }
	if err = json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	return policy.Queues
}

func assertRefused(t *testing.T, conn *rabbit.Connection, run func(*rabbit.Channel) error) {
	t.Helper()
	channel := permissionChannel(t, conn)
	closed := channel.NotifyClose(make(chan *rabbit.Error, 1))
	err := run(channel)
	if err == nil {
		select {
		case err = <-closed:
		case <-time.After(3 * time.Second):
			t.Fatal("operation was allowed; expected ACCESS_REFUSED")
		}
	}
	var denied *rabbit.Error
	if !errors.As(err, &denied) || denied == nil || denied.Code != 403 {
		t.Fatalf("expected ACCESS_REFUSED, got %v", err)
	}
}

func forbiddenPublish(exchange, key string) func(*rabbit.Channel) error {
	return func(channel *rabbit.Channel) error {
		return channel.PublishWithContext(context.Background(), exchange, key, false, false, rabbit.Publishing{Body: []byte("forbidden")})
	}
}

func assertProbe(t *testing.T, channel *rabbit.Channel, queue, marker string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		delivery, ok, err := channel.Get(queue, false)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			if string(delivery.Body) != marker {
				_ = delivery.Nack(false, true)
				t.Fatalf("unexpected delivery in %s", queue)
			}
			if err = delivery.Ack(false); err != nil {
				t.Fatal(err)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("probe unavailable in %s", queue)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestRuntimeBrokerCredentialsUseOnlyCatalogueQueues(t *testing.T) {
	admin := permissionChannel(t, permissionConnection(t, "BROKER_ADMIN_URL"))
	for owner, queues := range permissionQueues(t) {
		if owner == "api" {
			continue
		}
		t.Run(owner, func(t *testing.T) {
			url := os.Getenv(strings.ToUpper(owner) + "_BROKER_URL")
			conn := permissionConnection(t, strings.ToUpper(owner)+"_BROKER_URL")
			reader := permissionChannel(t, conn)
			for _, queue := range queues {
				delivery, ok, err := reader.Get(queue, false)
				if err != nil {
					t.Fatalf("catalogue queue %s: %v", queue, err)
				}
				if ok {
					if err = delivery.Nack(false, true); err != nil {
						t.Fatal(err)
					}
				}
			}
			publisher, err := Connect(url)
			if err != nil {
				t.Fatal(err)
			}
			defer publisher.Close()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			// An administrator owns the observer. Owners cannot consume arbitrary new queues.
			observer, err := admin.QueueDeclare("", false, true, true, false, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, visibility := range []string{"domain", "integration"} {
				key := visibility + "." + owner + ".permission-probe"
				if err = admin.QueueBind(observer.Name, key, Exchange, false, nil); err != nil {
					t.Fatal(err)
				}
				if err = publisher.send(ctx, Exchange, key, rabbit.Publishing{Body: []byte(key)}); err != nil {
					t.Fatal(err)
				}
				assertProbe(t, admin, observer.Name, key)
			}
			queue := ""
			for _, name := range queues {
				if strings.HasSuffix(name, ".dead") {
					queue = name
					break
				}
			}
			if queue == "" {
				t.Fatal("owner has no dead-letter queue")
			}
			if err = publisher.send(ctx, DeliveryExchange(owner), queue, rabbit.Publishing{Body: []byte("delivery-probe")}); err != nil {
				t.Fatal(err)
			}
			assertProbe(t, reader, queue, "delivery-probe")
			foreign := "menu"
			if owner == foreign {
				foreign = "ordering"
			}
			for _, probe := range []struct {
				name string
				run  func(*rabbit.Channel) error
			}{
				{"declare exchange", func(c *rabbit.Channel) error {
					return c.ExchangeDeclare(Exchange, "topic", true, false, false, false, nil)
				}},
				{"delete exchange", func(c *rabbit.Channel) error { return c.ExchangeDelete(Exchange, true, false) }},
				{"declare queue", func(c *rabbit.Channel) error {
					_, err := c.QueueDeclare("ref."+owner+".unregistered", false, false, false, false, nil)
					return err
				}},
				{"delete queue", func(c *rabbit.Channel) error { _, err := c.QueueDelete(queue, false, false, false); return err }},
				{"bind queue", func(c *rabbit.Channel) error { return c.QueueBind(queue, queue, DeliveryExchange(owner), false, nil) }},
				{"unbind queue", func(c *rabbit.Channel) error { return c.QueueUnbind(queue, queue, DeliveryExchange(owner), nil) }},
				{"bind exchange", func(c *rabbit.Channel) error {
					return c.ExchangeBind(DeliveryExchange(owner), "#", Exchange, false, nil)
				}},
				{"bind own delivery exchange", func(c *rabbit.Channel) error {
					return c.ExchangeBind(DeliveryExchange(owner), "#", DeliveryExchange(owner), false, nil)
				}},
				{"read unregistered owner queue", func(c *rabbit.Channel) error { _, _, err := c.Get("ref."+owner+".unregistered", false); return err }},
				{"read foreign queue", func(c *rabbit.Channel) error { _, _, err := c.Get(RequestQueue(foreign), false); return err }},
				{"publish foreign private event", forbiddenPublish(Exchange, "domain."+foreign+".permission-probe")},
				{"publish foreign integration event", forbiddenPublish(Exchange, "integration."+foreign+".permission-probe")},
				{"publish foreign delivery", forbiddenPublish(DeliveryExchange(foreign), RequestQueue(foreign)+".dead")},
				{"default exchange bypass", forbiddenPublish("", queue)},
			} {
				t.Run(probe.name, func(t *testing.T) { assertRefused(t, conn, probe.run) })
			}
		})
	}
}
