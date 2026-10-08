package requests

import (
	"context"
	"errors"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/shared"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"testing"
)

const id = "11111111-1111-4111-8111-111111111111"

func priceRequest() *pb.Request {
	zero := uint64(0)
	price := int64(0)
	code := "coffee"
	return &pb.Request{ContractVersion: 1, RequestId: id, Context: "menu", Payload: &pb.Request_Command{Command: &pb.Command{Metadata: &shared.CommandMetadata{CommandId: id, AggregateId: id, ExpectedVersion: &zero, CorrelationId: id}, Payload: &pb.Command_ChangePrice{ChangePrice: &pb.ChangePrice{Code: &code, Minor: &price}}}}}
}
func TestCommandTranslationRetainsPlainReceiptMaterialAndZero(t *testing.T) {
	type input struct {
		Code  string `json:"code"`
		Minor int64  `json:"minor"`
	}
	registry := New("menu")
	Command(registry, "changePrice", func(p *pb.ChangePrice) input { return input{Code: p.GetCode(), Minor: p.GetMinor()} }, func(_ context.Context, m a.Metadata, c input) (a.Outcome, error) {
		if c.Code != "coffee" || c.Minor != 0 || m.Input != c || m.Name != "menu.ChangePrice" || m.ID != id || m.ExpectedVersion == nil || *m.ExpectedVersion != 0 {
			t.Fatal("application command material changed")
		}
		if _, wire := m.Input.(proto.Message); wire {
			t.Fatal("generated message reached application")
		}
		return a.Outcome{AggregateID: id, Version: 1, Status: "draft"}, nil
	})
	reply := registry.Handle(t.Context(), priceRequest())
	if reply.GetOutcome() == nil || reply.GetOutcome().Version != 1 {
		t.Fatal("command reply failed")
	}
}
func TestInvalidOwnerAndMissingPriceNeverReachHandler(t *testing.T) {
	registry := New("menu")
	called := false
	Command(registry, "changePrice", func(*pb.ChangePrice) struct{} { return struct{}{} }, func(context.Context, a.Metadata, struct{}) (a.Outcome, error) { called = true; return a.Outcome{}, nil })
	request := priceRequest()
	request.GetCommand().GetChangePrice().Minor = nil
	if _, _, err := Validate(request, "menu"); err == nil {
		t.Fatal("missing price accepted")
	}
	if registry.Handle(t.Context(), request).GetError().Code != "invalid_request" || called {
		t.Fatal("invalid command reached application")
	}
	if _, _, err := Validate(priceRequest(), "ordering"); err == nil {
		t.Fatal("foreign request accepted")
	}
}
func TestInfrastructureFailureProducesRetryableReply(t *testing.T) {
	registry := New("menu")
	Command(registry, "changePrice", func(*pb.ChangePrice) struct{} { return struct{}{} }, func(context.Context, a.Metadata, struct{}) (a.Outcome, error) {
		return a.Outcome{}, errors.New("database credential private-secret")
	})
	reply := registry.Handle(t.Context(), priceRequest())
	if reply.GetError().Code != "temporarily_unavailable" || reply.GetError().Message != "" {
		t.Fatal("infrastructure failure leaked or became rejection")
	}
}
