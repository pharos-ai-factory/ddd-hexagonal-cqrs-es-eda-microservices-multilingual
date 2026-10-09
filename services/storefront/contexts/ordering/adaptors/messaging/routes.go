package messaging

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/integrations/requests"
)

// OrderingRequestHandlers groups the explicitly typed handlers mounted by this transport adaptor.
type OrderingRequestHandlers struct {
	Queries        app.OrderingQueries
	CreateOrder    app.CreateOrderCommandHandler
	AddLine        app.AddLineCommandHandler
	ChangeQuantity app.ChangeQuantityCommandHandler
	PlaceOrder     app.PlaceOrderCommandHandler
}

func Bind(h OrderingRequestHandlers) *rpc.RabbitMQRequestRegistry {
	r := rpc.New("ordering")
	rpc.Command(r, "createOrder", func(p *pb.CreateOrder) app.CreateOrderCommand {
		return app.CreateOrderCommand{CustomerID: p.GetCustomerId(), EditionID: p.GetEditionId()}
	}, h.CreateOrder.Execute)
	rpc.Command(r, "addLine", func(p *pb.AddLine) app.AddLineCommand {
		return app.AddLineCommand{LineID: p.GetLineId(), EditionID: p.GetEditionId(), OfferCode: p.GetOfferCode(), Quantity: int(p.GetQuantity())}
	}, h.AddLine.Execute)
	rpc.Command(r, "changeQuantity", func(p *pb.ChangeQuantity) app.ChangeQuantityCommand {
		return app.ChangeQuantityCommand{LineID: p.GetLineId(), Quantity: int(p.GetQuantity())}
	}, h.ChangeQuantity.Execute)
	rpc.Command(r, "placeOrder", func(p *pb.PlaceOrder) app.PlaceOrderCommand { return app.PlaceOrderCommand{} }, h.PlaceOrder.Execute)
	rpc.Queries(r, "order", "orders", h.Queries.List, h.Queries.Get, h.Queries.Page, orderReply, ordersReply)
	return r
}
