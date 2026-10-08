package messaging

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/integrations/requests"
)

type Handlers struct {
	Queries        app.OrderingQueries
	CreateOrder    app.CreateOrderHandler
	AddLine        app.AddLineHandler
	ChangeQuantity app.ChangeQuantityHandler
	PlaceOrder     app.PlaceOrderHandler
}

func Bind(h Handlers) *rpc.Registry {
	r := rpc.New("ordering")
	rpc.Command(r, "createOrder", func(p *pb.CreateOrder) app.CreateOrder {
		return app.CreateOrder{CustomerID: p.GetCustomerId(), EditionID: p.GetEditionId()}
	}, h.CreateOrder.Execute)
	rpc.Command(r, "addLine", func(p *pb.AddLine) app.AddLine {
		return app.AddLine{LineID: p.GetLineId(), EditionID: p.GetEditionId(), OfferCode: p.GetOfferCode(), Quantity: int(p.GetQuantity())}
	}, h.AddLine.Execute)
	rpc.Command(r, "changeQuantity", func(p *pb.ChangeQuantity) app.ChangeQuantity {
		return app.ChangeQuantity{LineID: p.GetLineId(), Quantity: int(p.GetQuantity())}
	}, h.ChangeQuantity.Execute)
	rpc.Command(r, "placeOrder", func(p *pb.PlaceOrder) struct{} { return struct{}{} }, h.PlaceOrder.Execute)
	rpc.Queries(r, "order", "orders", h.Queries.List, h.Queries.Get, h.Queries.Page, orderReply, ordersReply)
	return r
}
