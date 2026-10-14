//go:build integration

package amqp

import (
	"context"
	"errors"
	rabbit "github.com/rabbitmq/amqp091-go"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRequestReplyPermissionsKeepRuntimeAuthoritySeparate(t *testing.T) {
	if !strings.HasPrefix(os.Getenv("CAFE_DISPOSABLE_PROJECT"), "cafe-reference-test-") {
		t.Fatal("disposable infrastructure required")
	}
	for _, probe := range []struct{ name, url, exchange, key, queue string }{
		{"API cannot publish events", "API_BROKER_URL", Exchange, "integration.menu.fake", ""},
		{"API cannot forge replies", "API_BROKER_URL", RepliesExchange, "reply.menu", ""},
		{"API cannot read owner requests", "API_BROKER_URL", "", "", RequestQueue("menu")},
		{"owner cannot request commands", "MENU_BROKER_URL", RequestsExchange, "request.menu.command", ""},
		{"owner cannot forge another reply", "MENU_BROKER_URL", RepliesExchange, "reply.ordering", ""},
		{"owner cannot read another request", "MENU_BROKER_URL", "", "", RequestQueue("ordering")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			conn, err := rabbit.Dial(os.Getenv(probe.url))
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			channel, err := conn.Channel()
			if err != nil {
				t.Fatal(err)
			}
			defer channel.Close()
			closed := channel.NotifyClose(make(chan *rabbit.Error, 1))
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			if probe.queue != "" {
				_, _, err = channel.Get(probe.queue, false)
			} else {
				err = channel.PublishWithContext(ctx, probe.exchange, probe.key, false, false, rabbit.Publishing{Body: []byte("forbidden")})
			}
			if err == nil {
				select {
				case err = <-closed:
				case <-ctx.Done():
					err = ctx.Err()
				}
			}
			var denied *rabbit.Error
			if !errors.As(err, &denied) || denied.Code != 403 {
				t.Fatalf("expected ACCESS_REFUSED, got %v", err)
			}
		})
	}
}
