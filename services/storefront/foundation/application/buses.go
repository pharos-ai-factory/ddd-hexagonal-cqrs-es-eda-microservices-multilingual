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

// RealtimeTargetKind identifies an application-defined recipient category.
type RealtimeTargetKind string

const (
	RealtimeContextTarget   RealtimeTargetKind = "context"
	RealtimePrincipalTarget RealtimeTargetKind = "principal"
)

// RealtimeTarget is an application recipient resolved to a transport channel by an outer adaptor.
type RealtimeTarget struct {
	Kind RealtimeTargetKind `json:"kind"`
	ID   string             `json:"id"`
}

func ContextTarget(owner string) RealtimeTarget {
	return RealtimeTarget{Kind: RealtimeContextTarget, ID: owner}
}

func PrincipalTarget(principalID string) RealtimeTarget {
	return RealtimeTarget{Kind: RealtimePrincipalTarget, ID: principalID}
}

// RealtimeMessage selects a typed delivery target and one authoritative payload.
// The owner transaction supplies command identity and committed revision.
type RealtimeMessage struct {
	Target                   RealtimeTarget
	Source                   *Metadata
	Revision                 uint64
	Snapshot                 any
	CommandID, CorrelationID string
	Rejection                *domain.Violation
}

func SnapshotMessage[S any](target RealtimeTarget, snapshot S) RealtimeMessage {
	return RealtimeMessage{Target: target, Snapshot: snapshot}
}

func RejectionMessage(target RealtimeTarget, rejection domain.Violation) RealtimeMessage {
	return RealtimeMessage{Target: target, Rejection: &rejection}
}
