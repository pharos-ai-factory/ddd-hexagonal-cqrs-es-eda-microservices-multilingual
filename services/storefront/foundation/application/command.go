// Package application defines provider-independent command execution contracts.
package application

import (
	"context"
	"time"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

type Visibility string

const (
	Private Visibility = "domain"
	Public  Visibility = "integration"
)

// Publication is selected explicitly by a context's application mapper.
type Publication struct {
	Name       string
	Visibility Visibility
	Payload    any
}

// Metadata identifies a command attempt and preserves source receipt material.
type Metadata struct {
	ID              string
	AggregateID     string
	Name            string
	ExpectedVersion *uint64
	CorrelationID   string
	CausationID     string
	// Input is a typed command DTO, used only to fingerprint material input.
	Input         any
	Consumer      string
	SourceEventID string
	SourceHash    string
}

// Outcome records a committed aggregate revision or typed business rejection.
type Outcome struct {
	AggregateID string            `json:"aggregateId"`
	Version     uint64            `json:"version"`
	Status      string            `json:"status"`
	Rejection   *domain.Violation `json:"rejection,omitempty"`
}

// ApplicationError represents an expected command failure outside aggregate policy.
type ApplicationError struct {
	Code    string
	Message string
}

var _ domain.ExpectedError = (*ApplicationError)(nil)

func (e *ApplicationError) Error() string { return e.Code + ": " + e.Message }
func (e *ApplicationError) Rejection() domain.Violation {
	return domain.Violation{Code: e.Code, Message: e.Message}
}

// Loaded carries restored state and its persisted aggregate revision.
type Loaded[S any] struct {
	Exists  bool   `json:"exists"`
	Version uint64 `json:"version"`
	State   S      `json:"state"`
}

// Mutation describes one aggregate change and the events to persist atomically.
type Mutation[S any] struct {
	State        S
	Changed      bool
	Status       string
	Publications []Publication
}

// AggregateCommandPort is permanently bound to one context and aggregate kind.
// It exposes neither a database transaction nor a second repository.
type AggregateCommandPort[S any] interface {
	Execute(context.Context, Metadata, func(Loaded[S]) (Mutation[S], error)) (Outcome, error)
}

type QueryPort[S any] interface {
	Get(context.Context, string) (Loaded[S], error)
	List(context.Context) ([]Loaded[S], error)
}

type Clock func() time.Time
type IDFactory func(purpose, stableKey string) string

func Changed[S any](state S, status string, publications ...Publication) Mutation[S] {
	return Mutation[S]{State: state, Changed: true, Status: status, Publications: publications}
}

func Expected(version uint64) *uint64 { return &version }
