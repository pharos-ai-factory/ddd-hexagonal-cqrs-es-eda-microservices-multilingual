package amqp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	rabbit "github.com/rabbitmq/amqp091-go"
)

var replyInstance = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type replyRouteKey struct{}

// ReplyRoute carries validated transport metadata through dispatch to reply persistence.
// An absent address retains the published route for older callers.
func ReplyRoute(ctx context.Context) string {
	route, _ := ctx.Value(replyRouteKey{}).(string)
	return route
}

func withReplyRoute(ctx context.Context, owner, route string) context.Context {
	// Existing durable replies predate explicit addresses and store the empty
	// legacy destination. Redelivery must compare equal after an upgrade.
	if route == "reply."+owner {
		route = ""
	}
	return context.WithValue(ctx, replyRouteKey{}, route)
}

func validateReplyRoute(owner, route string) (string, error) {
	legacy := "reply." + owner
	if route == "" || route == legacy {
		return legacy, nil
	}
	instance, found := strings.CutPrefix(route, legacy+".")
	if !found || !replyInstance.MatchString(instance) {
		return "", fmt.Errorf("invalid owner reply address")
	}
	return route, nil
}

func (p *ConfirmedPublisher) PublishReply(ctx context.Context, owner, id, route string, body []byte) error {
	route, err := validateReplyRoute(owner, route)
	if err != nil {
		return err
	}
	return p.send(ctx, RepliesExchange, route, rabbit.Publishing{ContentType: "application/x-protobuf", DeliveryMode: rabbit.Persistent, MessageId: id, CorrelationId: id, Type: "reply", AppId: owner, Body: body})
}
