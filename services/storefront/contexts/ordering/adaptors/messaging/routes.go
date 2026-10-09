package messaging

import (
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/queries"
	orderingcommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/integrations/requests"
)

// OrderingRequestHandlers groups the explicitly typed handlers mounted by this transport adaptor.
type OrderingRequestHandlers struct {
	Queries        endpoints.OrderQueryEndpoints
	CreateOrder    execution.Executor[orderingcommands.CreateOrderCommand]
	AddLine        execution.Executor[orderingcommands.AddLineCommand]
	ChangeQuantity execution.Executor[orderingcommands.ChangeQuantityCommand]
	PlaceOrder     execution.Executor[orderingcommands.PlaceOrderCommand]
}

func Bind(h OrderingRequestHandlers) *rpc.RabbitMQRequestRegistry {
	r := rpc.New("ordering")
	rpc.Command(r, "createOrder", func(p *pb.CreateOrder) orderingcommands.CreateOrderCommand {
		return orderingcommands.CreateOrderCommand{CustomerID: p.GetCustomerId(), EditionID: p.GetEditionId()}
	}, h.CreateOrder.Execute)
	rpc.Command(r, "addLine", func(p *pb.AddLine) orderingcommands.AddLineCommand {
		return orderingcommands.AddLineCommand{LineID: p.GetLineId(), EditionID: p.GetEditionId(), OfferCode: p.GetOfferCode(), Quantity: int(p.GetQuantity())}
	}, h.AddLine.Execute)
	rpc.Command(r, "changeQuantity", func(p *pb.ChangeQuantity) orderingcommands.ChangeQuantityCommand {
		return orderingcommands.ChangeQuantityCommand{LineID: p.GetLineId(), Quantity: int(p.GetQuantity())}
	}, h.ChangeQuantity.Execute)
	rpc.Command(r, "placeOrder", func(p *pb.PlaceOrder) orderingcommands.PlaceOrderCommand { return orderingcommands.PlaceOrderCommand{} }, h.PlaceOrder.Execute)
	rpc.Queries(r, "order", "orders", h.Queries.List, h.Queries.Get, h.Queries.Page, orderReply, ordersReply)
	return r
}
