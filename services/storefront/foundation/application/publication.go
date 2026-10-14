package application

import "context"

// DurablePublicationPort atomically records this subscriber's receipt and outgoing integration bytes.
type DurablePublicationPort[P any] interface {
	Publish(context.Context, Metadata, P) (Outcome, error)
}
