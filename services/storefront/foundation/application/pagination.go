package application

import (
	"context"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/shared"
)

// PageRequest opts into a bounded keyset scan. List remains an unpaginated query.
type PageRequest = shared.PageRequest

// Page contains a bounded set of versioned read values and a continuation identity.
type Page[S any] struct {
	Items  []Loaded[S]
	NextID string
}
type PagedQueryPort[S any] interface {
	QueryPort[S]
	Page(context.Context, PageRequest) (Page[S], error)
}
