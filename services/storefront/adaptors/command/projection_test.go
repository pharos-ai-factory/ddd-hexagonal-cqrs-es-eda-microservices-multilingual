package command

import (
	"context"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"testing"
)

type projectionFixture struct {
	t       *testing.T
	active  *bool
	calls   int
	failure bool
}

func (p *projectionFixture) Find(context.Context, string) (string, bool, error) {
	p.t.Helper()
	if *p.active {
		p.t.Fatal("projection acquired a connection during the transaction")
	}
	p.calls++
	if p.failure {
		return "", false, fmt.Errorf("projection unavailable")
	}
	return "published", true, nil
}
func (p *projectionFixture) Record(context.Context, a.Metadata, string, uint64, string) error {
	p.t.Fatal("unexpected projection write")
	return nil
}

type transactionFixture struct {
	active *bool
	calls  int
}

func (p *transactionFixture) Execute(ctx context.Context, m a.Metadata, work func(a.WriteRepository[string]) (Result, error)) (a.Outcome, error) {
	p.calls++
	*p.active = true
	defer func() { *p.active = false }()
	result, err := work(nil)
	return a.Outcome{AggregateID: m.AggregateID, Status: result.Status}, err
}

type projectionCommandHandler struct {
	t          *testing.T
	active     *bool
	projection a.ProjectionPort[string]
}

func (h projectionCommandHandler) Execute(ctx context.Context, _ a.CommandContext, c string) (a.CommandResult, error) {
	if !*h.active {
		h.t.Fatal("handler escaped command transaction")
	}
	value, found, err := h.projection.Find(ctx, c)
	if err != nil || !found {
		return a.CommandResult{}, fmt.Errorf("prepared projection lost: %v", err)
	}
	return a.Result(value), nil
}
func TestProjectionIsPreparedBeforeTransaction(t *testing.T) {
	active := false
	source := &projectionFixture{t: t, active: &active}
	transaction := &transactionFixture{active: &active}
	executor := BindProjection(transaction, source, func(c string) string { return c }, func(_ a.WriteRepository[string], projection a.ProjectionPort[string]) projectionCommandHandler {
		return projectionCommandHandler{t, &active, projection}
	})
	outcome, err := executor.Execute(t.Context(), a.Metadata{AggregateID: "target"}, "revision")
	if err != nil || outcome.Status != "published" || source.calls != 1 || transaction.calls != 1 {
		t.Fatalf("execution: %+v %v, reads=%d transactions=%d", outcome, err, source.calls, transaction.calls)
	}
	source.failure = true
	if _, err = executor.Execute(t.Context(), a.Metadata{}, "revision"); err == nil || transaction.calls != 1 {
		t.Fatal("failed preflight entered transaction")
	}
}
