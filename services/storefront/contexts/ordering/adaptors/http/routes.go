package http

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	"net/http"
)

// OrderingHTTPHandlers groups the explicitly typed handlers mounted by this transport adaptor.
type OrderingHTTPHandlers struct {
	OrderingQueries app.OrderingQueries
	CreateOrder     app.CreateOrderCommandHandler
	AddLine         app.AddLineCommandHandler
	ChangeQuantity  app.ChangeQuantityCommandHandler
	PlaceOrder      app.PlaceOrderCommandHandler
}

func Mount(mux interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}, h OrderingHTTPHandlers) {
	mux.HandleFunc("GET /v1/ordering/orders", web.PagedList(h.OrderingQueries.List, h.OrderingQueries.Page))
	mux.HandleFunc("GET /v1/ordering/orders/{id}", web.Get(h.OrderingQueries.Get))
	mux.HandleFunc("POST /v1/ordering/orders/{id}", web.Command("ordering.CreateOrder", h.CreateOrder.Execute))
	mux.HandleFunc("POST /v1/ordering/orders/{id}/lines", web.Command("ordering.AddLine", h.AddLine.Execute))
	mux.HandleFunc("POST /v1/ordering/orders/{id}/quantities", web.Command("ordering.ChangeQuantity", h.ChangeQuantity.Execute))
	mux.HandleFunc("POST /v1/ordering/orders/{id}/place", web.Command("ordering.PlaceOrder", h.PlaceOrder.Execute))
}
