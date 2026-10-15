package di

import (
	"context"
	"errors"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/shared"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/tests/support"
	"google.golang.org/protobuf/proto"
	"testing"
)

type queryReader struct {
	failed bool
	page   a.PageRequest
	calls  int
}

func (r *queryReader) Get(_ context.Context, id string) (a.Loaded[view.DrinkView], error) {
	r.calls++
	if r.failed {
		return a.Loaded[view.DrinkView]{}, errors.New("private database detail")
	}
	if id != support.Drink {
		return a.Loaded[view.DrinkView]{}, nil
	}
	return a.Loaded[view.DrinkView]{Exists: true, Version: 7, State: view.DrinkView{ID: id, Name: "Coffee"}}, nil
}
func (r *queryReader) List(context.Context) ([]a.Loaded[view.DrinkView], error) {
	r.calls++
	return make([]a.Loaded[view.DrinkView], 101), nil
}
func (r *queryReader) Page(_ context.Context, page a.PageRequest) (a.Page[view.DrinkView], error) {
	r.calls++
	r.page = page
	return a.Page[view.DrinkView]{NextID: support.Drink}, nil
}

func TestDirectQueryRegistrationPreservesReadBehaviour(t *testing.T) {
	reader := &queryReader{}
	registry := menuRequestsProvider(nil, nil, nil, q.GetDrinkQueryHandler{ReadRepository: reader}, q.ListDrinksQueryHandler{ReadRepository: reader}, q.GetEditionQueryHandler{}, q.ListEditionsQueryHandler{})
	request := &pb.Request{ContractVersion: 1, RequestId: support.Customer, Context: "menu", Payload: &pb.Request_Query{Query: &pb.Query{Payload: &pb.Query_GetDrink{GetDrink: &pb.GetDrink{Id: support.Drink}}}}}
	reply := registry.Handle(t.Context(), request).(*pb.Reply)
	if reply.GetDrink().Version != 7 || reply.GetDrink().State.Id != support.Drink {
		t.Fatalf("item: %v", reply)
	}
	request.GetQuery().GetGetDrink().Id = support.Customer
	if registry.Handle(t.Context(), request).GetError().Code != "not_found" {
		t.Fatal("missing root lost")
	}
	request.GetQuery().Payload = &pb.Query_ListDrinks{ListDrinks: &pb.ListDrinks{}}
	reply = registry.Handle(t.Context(), request).(*pb.Reply)
	if len(reply.GetDrinks().Items) != 101 || reply.GetDrinks().Paged {
		t.Fatal("complete read truncated")
	}
	request.GetQuery().GetListDrinks().Page = &shared.PageRequest{Limit: 2, After: proto.String(support.Drink)}
	reply = registry.Handle(t.Context(), request).(*pb.Reply)
	if !reply.GetDrinks().Paged || reply.GetDrinks().GetNextId() != support.Drink || reader.page.Limit != 2 || reader.page.After == nil || *reader.page.After != support.Drink {
		t.Fatal("pagination changed")
	}
	before := reader.calls
	request.GetQuery().GetListDrinks().Page.Limit = 0
	if registry.Handle(t.Context(), request).GetError().Code != "invalid_request" || reader.calls != before {
		t.Fatal("invalid page reached handler")
	}
	reader.failed = true
	request.GetQuery().Payload = &pb.Query_GetDrink{GetDrink: &pb.GetDrink{Id: support.Drink}}
	reply = registry.Handle(t.Context(), request).(*pb.Reply)
	if reply.GetError().Code != "temporarily_unavailable" || reply.GetError().Message != "" {
		t.Fatal("storage failure lost or disclosed")
	}
}
