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

type Outcome struct {
	AggregateID string            `json:"aggregateId"`
	Version     uint64            `json:"version"`
	Status      string            `json:"status"`
	Rejection   *domain.Violation `json:"rejection,omitempty"`
}

type Loaded[S any] struct {
	Exists  bool   `json:"exists"`
	Version uint64 `json:"version"`
	State   S      `json:"state"`
}

type Mutation[S any] struct {
	State        S
	Changed      bool
	Status       string
	Publications []Publication
}

// CommandPort is permanently bound to one context and aggregate kind.
// It exposes neither a database transaction nor a second repository.
type CommandPort[S any] interface {
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

func MapFacts(facts []domain.Fact, mapper func(domain.Fact) []Publication) []Publication {
	var result []Publication
	for _, fact := range facts {
		result = append(result, mapper(fact)...)
	}
	return result
}
