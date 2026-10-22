package command

import (
	"context"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"testing"
)

type actorTransaction struct{}

func (actorTransaction) Execute(ctx context.Context, m a.Metadata, work func(a.WriteRepository[int]) (Result, error)) (a.Outcome, error) {
	result, err := work(nil)
	return a.Outcome{AggregateID: m.AggregateID, Status: result.Status}, err
}

type actorHandler struct {
	t        *testing.T
	expected a.CommandContext
}

func (h actorHandler) Execute(_ context.Context, c a.CommandContext, _ int) (a.CommandResult, error) {
	if c != h.expected {
		h.t.Fatalf("handler context: %+v", c)
	}
	return a.Result("accepted"), nil
}
func TestHandlerReceivesAuthenticatedActor(t *testing.T) {
	for _, actor := range []string{"", "11111111-1111-4111-8111-111111111111"} {
		expected := a.CommandContext{Target: "22222222-2222-4222-8222-222222222222", ActorID: actor}
		executor := Bind[int, int](actorTransaction{}, func(a.WriteRepository[int], a.RealtimeBusPort[a.RealtimeMessage]) actorHandler {
			return actorHandler{t, expected}
		})
		result, err := executor.Execute(t.Context(), a.Metadata{AggregateID: expected.Target, ActorID: actor}, 0)
		if err != nil || result.Status != "accepted" {
			t.Fatalf("%+v %v", result, err)
		}
	}
}
