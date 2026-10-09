package application

import "context"

// PageRequest opts into a bounded keyset scan. List remains an unpaginated query.
type PageRequest struct {
	Limit int
	After string
}

// Page contains a bounded set of versioned read values and a continuation identity.
type Page[S any] struct {
	Items  []Loaded[S]
	NextID string
}
type PagedQueryPort[S any] interface {
	QueryPort[S]
	Page(context.Context, PageRequest) (Page[S], error)
}
