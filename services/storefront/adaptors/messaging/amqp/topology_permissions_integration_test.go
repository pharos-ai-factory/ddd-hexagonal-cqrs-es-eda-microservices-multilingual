//go:build integration

package amqp

import (
	"os"
	"testing"

	rabbit "github.com/rabbitmq/amqp091-go"
)

func TestTopologyCredentialDeclaresCatalogueWithoutAdministratorAuthority(t *testing.T) {
	// Fresh Compose bootstrap already declared every request and event with this credential.
	// Repeating representative private/command declarations also proves idempotency and DLX authority.
	url := os.Getenv("TOPOLOGY_BROKER_URL")
	conn := permissionConnection(t, "TOPOLOGY_BROKER_URL")
	if err := DeclareRequests(url, []string{"menu"}); err != nil {
		t.Fatal(err)
	}
	if err := Declare(url, []Binding{{Consumer: "loyalty.issue-reward", Context: "loyalty", Event: "loyalty.reward-earned", Visibility: "domain", Command: true}}); err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct {
		name string
		run  func(*rabbit.Channel) error
	}{
		{"unregistered queue", func(c *rabbit.Channel) error {
			_, err := c.QueueDeclare("ref.menu.unregistered", true, false, false, false, nil)
			return err
		}},
		{"unregistered exchange", func(c *rabbit.Channel) error {
			return c.ExchangeDeclare("ref.foreign.delivery", "direct", true, false, false, false, nil)
		}},
		{"unregistered event binding", func(c *rabbit.Channel) error {
			return c.QueueBind("ref.menu.drink-directory", "domain.menu.unregistered", Exchange, false, nil)
		}},
		{"wildcard binding", func(c *rabbit.Channel) error {
			return c.QueueBind("ref.menu.drink-directory", "#", Exchange, false, nil)
		}},
		{"unregistered request binding", func(c *rabbit.Channel) error {
			return c.QueueBind(RequestQueue("menu"), "request.foreign.command", RequestsExchange, false, nil)
		}},
		{"unregistered reply binding", func(c *rabbit.Channel) error {
			return c.QueueBind(ReplyQueue("menu"), "reply.foreign", RepliesExchange, false, nil)
		}},
		{"read private event queue", func(c *rabbit.Channel) error { _, _, err := c.Get("ref.menu.drink-directory", false); return err }},
		{"read API replies", func(c *rabbit.Channel) error { _, _, err := c.Get(ReplyQueue("menu"), false); return err }},
		{"publish events", forbiddenPublish(Exchange, "domain.menu.drink-published")},
		{"publish requests", forbiddenPublish(RequestsExchange, "request.menu.command")},
		{"publish replies", forbiddenPublish(RepliesExchange, "reply.menu")},
	} {
		t.Run(probe.name, func(t *testing.T) { assertRefused(t, conn, probe.run) })
	}
}
