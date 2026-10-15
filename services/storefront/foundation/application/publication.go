package application

import "context"

// IntegrationEventBusPort atomically records this subscriber's receipt and outgoing integration bytes.
type IntegrationEventBusPort[P any] interface {
	Publish(context.Context, Metadata, P) (Outcome, error)
}
