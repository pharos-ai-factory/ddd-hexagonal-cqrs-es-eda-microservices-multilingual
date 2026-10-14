package postgres

import (
	"context"
	"errors"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"testing"
)

type mappedSource struct {
	failure error
	request a.PageRequest
}

func (s *mappedSource) Get(_ context.Context, id string) (a.Loaded[int], error) {
	return a.Loaded[int]{Exists: id != "missing", Version: 7, State: 3}, s.failure
}
func (s *mappedSource) List(ctx context.Context) ([]a.Loaded[int], error) {
	v, e := s.Get(ctx, "present")
	return []a.Loaded[int]{v}, e
}
func (s *mappedSource) Page(ctx context.Context, p a.PageRequest) (a.Page[int], error) {
	s.request = p
	v, e := s.List(ctx)
	return a.Page[int]{Items: v, NextID: "next"}, e
}
func TestMappedReadRepositoryPreserveMissingRowsRevisionsAndFailures(t *testing.T) {
	ctx := context.Background()
	source := &mappedSource{}
	calls := 0
	q := MappedReadRepository[int, string]{Source: source, View: func(value int) (string, error) { calls++; return "read", nil }}
	missing, err := q.Get(ctx, "missing")
	if err != nil || missing.Exists || calls != 0 {
		t.Fatal("mapped a missing aggregate")
	}
	page, err := q.Page(ctx, a.PageRequest{Limit: 2, After: "after"})
	if err != nil || page.NextID != "next" || len(page.Items) != 1 || page.Items[0].Version != 7 || page.Items[0].State != "read" || source.request.After != "after" {
		t.Fatalf("page changed: %+v %v", page, err)
	}
	source.failure = errors.New("storage unavailable")
	if _, err = q.List(ctx); !errors.Is(err, source.failure) {
		t.Fatal("lost storage failure")
	}
	source.failure = nil
	invalid := errors.New("corrupt stored state")
	q.View = func(int) (string, error) { return "", invalid }
	if _, err = q.Page(ctx, a.PageRequest{Limit: 2}); !errors.Is(err, invalid) {
		t.Fatal("lost restoration failure")
	}
}
