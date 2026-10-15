package messaging

import (
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetOrderReply maps the resource revision or the existing missing-resource response.
func GetOrderReply(_ orderingcontracts.GetOrderQuery, loaded a.Loaded[view.OrderView]) envelope.Reply {
	if !loaded.Exists {
		return &pb.Reply{Payload: &pb.Reply_Error{Error: &envelope.RequestError{Code: "not_found"}}}
	}
	return &pb.Reply{Payload: &pb.Reply_Order{Order: orderReply(loaded).(*pb.LoadedOrder)}}
}

// ListOrdersReply retains pagination and its continuation from the handler result.
func ListOrdersReply(query orderingcontracts.ListOrdersQuery, page a.Page[view.OrderView]) envelope.Reply {
	return &pb.Reply{Payload: &pb.Reply_Orders{Orders: ordersReply(page.Items, query.Page != nil, page.NextID).(*pb.Orders)}}
}
