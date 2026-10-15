package messaging

import (
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
	model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/shared"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestGeneratedResponseEncodesNestedListsPresenceAndZero(t *testing.T) {
	empty := ""
	result := model.ListOrdersResponse{Orders: &model.Orders{Paged: true, NextID: &empty, Items: []model.LoadedOrder{{Exists: true, Version: 7, State: model.Order{ID: "order", Lines: []model.Line{{ID: "line", Quantity: 1, Selection: model.Selection{Name: "Coffee", Minor: 0}}}}}}}}
	wire, err := EncodeListOrdersResponse(model.ListOrdersQuery{}, result)
	if err != nil {
		t.Fatal(err)
	}
	body, err := proto.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	decoded := &pb.Reply{}
	if err = proto.Unmarshal(body, decoded); err != nil {
		t.Fatal(err)
	}
	orders := decoded.GetOrders()
	if orders.NextId == nil || *orders.NextId != "" || !orders.Paged || len(orders.Items) != 1 || orders.Items[0].Version != 7 || orders.Items[0].State.Lines[0].Selection.Minor != 0 {
		t.Fatal("response lost nested data, presence or revision")
	}
	result.Orders.NextID = nil
	wire, err = EncodeListOrdersResponse(model.ListOrdersQuery{}, result)
	if err != nil || wire.(*pb.Reply).GetOrders().NextId != nil {
		t.Fatal("absent continuation became present")
	}
}

func TestGeneratedResponseRejectsInvalidAlternativesAndOverflow(t *testing.T) {
	query := model.GetOrderQuery{}
	for _, value := range []model.GetOrderResponse{{}, {Order: &model.LoadedOrder{}, Error: &shared.RequestError{Code: "not_found"}}} {
		if _, err := EncodeGetOrderResponse(query, value); err == nil {
			t.Fatal("ambiguous response accepted")
		}
	}
	value := model.GetOrderResponse{Order: &model.LoadedOrder{State: model.Order{Lines: []model.Line{{Quantity: 1 << 31}}}}}
	if _, err := EncodeGetOrderResponse(query, value); err == nil {
		t.Fatal("quantity overflow accepted")
	}
	result, err := EncodeGetOrderResponse(query, model.GetOrderResponse{Error: &shared.RequestError{Code: "not_found"}})
	if err != nil || result.GetError().Code != "not_found" {
		t.Fatal("missing-resource error lost")
	}
}
