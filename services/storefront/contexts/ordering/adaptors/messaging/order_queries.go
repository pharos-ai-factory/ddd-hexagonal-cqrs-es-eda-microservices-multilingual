package messaging

import (
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/queries"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetOrder decodes the published item query into its application input.
func GetOrder(value *pb.GetOrder) q.GetOrderQuery { return q.GetOrderQuery{ID: value.GetId()} }

// ListOrders preserves complete reads and explicitly requested pagination.
func ListOrders(value *pb.ListOrders) q.ListOrdersQuery {
	result := q.ListOrdersQuery{}
	if page := value.GetPage(); page != nil {
		result.Page = &a.PageRequest{Limit: int(page.Limit), After: page.GetAfter()}
	}
	return result
}

// GetOrderReply maps the resource revision or the existing missing-resource response.
func GetOrderReply(_ q.GetOrderQuery, loaded a.Loaded[view.OrderView]) envelope.Reply {
	if !loaded.Exists {
		return &pb.Reply{Payload: &pb.Reply_Error{Error: &envelope.RequestError{Code: "not_found"}}}
	}
	return &pb.Reply{Payload: &pb.Reply_Order{Order: orderReply(loaded).(*pb.LoadedOrder)}}
}

// ListOrdersReply retains pagination and its continuation from the handler result.
func ListOrdersReply(query q.ListOrdersQuery, page a.Page[view.OrderView]) envelope.Reply {
	return &pb.Reply{Payload: &pb.Reply_Orders{Orders: ordersReply(page.Items, query.Page != nil, page.NextID).(*pb.Orders)}}
}
