package messaging

import (
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/requests"
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging/queryendpoints"
	commands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
)

// CreateOrder maps the published payload to the application command.
func CreateOrder(p *pb.CreateOrder) commands.CreateOrderCommand {
	return commands.CreateOrderCommand{CustomerID: p.GetCustomerId(), EditionID: p.GetEditionId()}

}

// AddLine maps the published payload to the application command.
func AddLine(p *pb.AddLine) commands.AddLineCommand {
	return commands.AddLineCommand{LineID: p.GetLineId(), EditionID: p.GetEditionId(), OfferCode: p.GetOfferCode(), Quantity: int(p.GetQuantity())}

}

// ChangeQuantity maps the published payload to the application command.
func ChangeQuantity(p *pb.ChangeQuantity) commands.ChangeQuantityCommand {
	return commands.ChangeQuantityCommand{LineID: p.GetLineId(), Quantity: int(p.GetQuantity())}

}

// PlaceOrder maps the published payload to the application command.
func PlaceOrder(p *pb.PlaceOrder) commands.PlaceOrderCommand {
	return commands.PlaceOrderCommand{}
}

// RegisterQueries mounts the context read endpoints.
func RegisterQueries(r *rpc.RabbitMQRequestRegistry, orders endpoints.OrderQueryEndpoints) {
	rpc.Queries(r, "order", "orders", orders.List, orders.Get, orders.Page, orderReply, ordersReply)
}
