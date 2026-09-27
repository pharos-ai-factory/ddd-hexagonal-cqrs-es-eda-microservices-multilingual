package application

import "context"

// ProjectionPort owns a consumer's local copy of immutable published facts.
// Projection writes do not grant authority over the producer's aggregates.
type ProjectionPort[S any] interface {
	Find(context.Context, string) (S, bool, error)
	Record(context.Context, Metadata, string, uint64, S) error
}

// DeliveryPort is an idempotent provider capability, never a provider SDK.
type DeliveryPort interface {
	Deliver(context.Context, Delivery) (string, error)
}

type Delivery struct {
	ID        string `json:"id"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}
