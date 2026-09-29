package http

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	"net/http"
)

type Handlers struct {
	OrderingQueries app.OrderingQueries
	CreateOrder     app.CreateOrderHandler
	AddLine         app.AddLineHandler
	ChangeQuantity  app.ChangeQuantityHandler
	PlaceOrder      app.PlaceOrderHandler
}

func Mount(mux *http.ServeMux, h Handlers) {
	mux.HandleFunc("GET /v1/ordering/orders", web.PagedList(h.OrderingQueries.List, h.OrderingQueries.Page))
	mux.HandleFunc("GET /v1/ordering/orders/{id}", web.Get(h.OrderingQueries.Get))
	mux.HandleFunc("POST /v1/ordering/orders/{id}", web.Command("ordering.CreateOrder", h.CreateOrder.Execute))
	mux.HandleFunc("POST /v1/ordering/orders/{id}/lines", web.Command("ordering.AddLine", h.AddLine.Execute))
	mux.HandleFunc("POST /v1/ordering/orders/{id}/quantities", web.Command("ordering.ChangeQuantity", h.ChangeQuantity.Execute))
	mux.HandleFunc("POST /v1/ordering/orders/{id}/place", web.Command("ordering.PlaceOrder", h.PlaceOrder.Execute))
}
