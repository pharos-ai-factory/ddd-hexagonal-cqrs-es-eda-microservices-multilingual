package application

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// CommandBusPort accepts durable owner commands; execution follows acceptance.
type CommandBusPort[C any] interface {
	Enqueue(context.Context, Metadata, C) (Outcome, error)
}

// QueryBusPort dispatches a typed query to its single registered handler.
type QueryBusPort[Q, R any] interface {
	Ask(context.Context, Q) (R, error)
}

// DomainEventBusPort stages private facts inside their owning transaction.
type DomainEventBusPort[E any] interface {
	Publish(context.Context, E) error
}

// RealtimeBusPort explicitly stages a browser message within the current unit of work.
// The surrounding transaction commits the message before delivery can begin.
type RealtimeBusPort[E any] interface {
	Publish(context.Context, E) error
}

// RealtimeEvent selects an authoritative snapshot or a committed operation rejection.
// Snapshot revision and audience are supplied by the owner transaction.
type RealtimeEvent struct {
	Source                   *Metadata
	Revision                 uint64
	Snapshot                 any
	CommandID, CorrelationID string
	Rejection                *domain.Violation
}
