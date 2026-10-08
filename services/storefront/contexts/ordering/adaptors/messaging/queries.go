package messaging

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
)

func orderReply(value a.Loaded[d.State]) proto.Message {
	state := &pb.Order{Id: value.State.ID, CustomerId: value.State.CustomerID, EditionId: value.State.EditionID, Currency: value.State.Currency, Status: value.State.Status, Lines: make([]*pb.Line, 0, len(value.State.Lines))}
	for _, line := range value.State.Lines {
		state.Lines = append(state.Lines, &pb.Line{Id: line.ID, Quantity: int32(line.Quantity), Selection: &pb.Selection{OfferCode: line.Selection.OfferCode, Name: line.Selection.Name, Minor: line.Selection.Minor}})
	}
	return &pb.LoadedOrder{Exists: value.Exists, Version: value.Version, State: state}
}
func ordersReply(values []a.Loaded[d.State], paged bool, next string) proto.Message {
	result := &pb.Orders{Items: make([]*pb.LoadedOrder, 0, len(values)), Paged: paged}
	for _, item := range values {
		result.Items = append(result.Items, orderReply(item).(*pb.LoadedOrder))
	}
	if next != "" {
		result.NextId = &next
	}
	return result
}
