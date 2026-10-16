//go:build integration

package amqp

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	rabbit "github.com/rabbitmq/amqp091-go"
)

func TestRequestReplyPermissionsCoverEveryOwner(t *testing.T) {
	api := permissionConnection(t, "API_BROKER_URL")
	reader := permissionChannel(t, api)
	publisher, err := Connect(os.Getenv("API_BROKER_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer publisher.Close()
	for owner := range permissionQueues(t) {
		if owner == "api" {
			continue
		}
		t.Run(owner, func(t *testing.T) {
			conn := permissionConnection(t, strings.ToUpper(owner)+"_BROKER_URL")
			ownerReader := permissionChannel(t, conn)
			replies, err := Connect(os.Getenv(strings.ToUpper(owner) + "_BROKER_URL"))
			if err != nil {
				t.Fatal(err)
			}
			defer replies.Close()
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			for _, kind := range []string{"command", "query"} {
				key := "request." + owner + "." + kind
				if err = publisher.send(ctx, RequestsExchange, key, rabbit.Publishing{Body: []byte(key)}); err != nil {
					t.Fatal(err)
				}
				assertProbe(t, ownerReader, RequestQueue(owner, kind), key)
			}
			if err = replies.PublishReply(ctx, owner, "permission-probe", "", []byte(owner)); err != nil {
				t.Fatal(err)
			}
			assertProbe(t, reader, ReplyQueue(owner), owner)
			assertRefused(t, conn, forbiddenPublish(RequestsExchange, "request."+owner+".command"))
			assertRefused(t, conn, func(c *rabbit.Channel) error { _, _, err := c.Get(ReplyQueue(owner), false); return err })
			foreign := "menu"
			if owner == foreign {
				foreign = "ordering"
			}
			assertRefused(t, conn, forbiddenPublish(RepliesExchange, "reply."+foreign))
		})
	}
	for _, probe := range []struct {
		name string
		run  func(*rabbit.Channel) error
	}{
		{"events", forbiddenPublish(Exchange, "integration.menu.fake")},
		{"replies", forbiddenPublish(RepliesExchange, "reply.menu")},
		{"unregistered request owner", forbiddenPublish(RequestsExchange, "request.foreign.command")},
		{"unregistered request kind", forbiddenPublish(RequestsExchange, "request.menu.other")},
		{"owner requests", func(c *rabbit.Channel) error { _, _, err := c.Get(RequestQueue("menu"), false); return err }},
		{"unregistered reply queue", func(c *rabbit.Channel) error { _, _, err := c.Get("ref.api.foreign.replies", false); return err }},
		{"bind reply queue", func(c *rabbit.Channel) error {
			return c.QueueBind(ReplyQueue("menu"), "reply.menu", RepliesExchange, false, nil)
		}},
		{"declare reply queue", func(c *rabbit.Channel) error {
			_, err := c.QueueDeclare(ReplyQueue("menu"), true, false, false, false, nil)
			return err
		}},
		{"delivery exchange", forbiddenPublish(DeliveryExchange("menu"), "ref.menu.drink-directory")},
		{"default exchange", forbiddenPublish("", RequestQueue("menu"))},
	} {
		t.Run(probe.name, func(t *testing.T) { assertRefused(t, api, probe.run) })
	}
}
