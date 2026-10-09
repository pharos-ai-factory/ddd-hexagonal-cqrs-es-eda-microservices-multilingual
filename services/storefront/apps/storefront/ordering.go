package main

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/http"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/postgres"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	"go.uber.org/fx"
)

func orderingModule() fx.Option {
	return fx.Module("ordering", fx.Provide(
		func(s *support.StorefrontRuntime) a.AggregateCommandPort[d.OrderState] {
			return pg.OrderCommands(s.Databases["ordering"])
		},
		func(s *support.StorefrontRuntime) a.PagedQueryPort[d.OrderState] {
			return pg.OrderQueries(s.Databases["ordering"])
		},
		func(s *support.StorefrontRuntime) a.ProjectionPort[model.MenuPublished] {
			return store.Project[model.MenuPublished](s.Databases["ordering"], "published-menus")
		},
		func(port a.AggregateCommandPort[d.OrderState], directory a.ProjectionPort[model.MenuPublished]) app.CreateOrderCommandHandler {
			return app.CreateOrderCommandHandler{Orders: port, Menus: directory}
		},
		func(port a.AggregateCommandPort[d.OrderState], directory a.ProjectionPort[model.MenuPublished]) app.AddLineCommandHandler {
			return app.AddLineCommandHandler{Orders: port, Menus: directory}
		},
		func(port a.AggregateCommandPort[d.OrderState]) app.ChangeQuantityCommandHandler {
			return app.ChangeQuantityCommandHandler{Orders: port}
		},
		func(port a.AggregateCommandPort[d.OrderState]) app.PlaceOrderCommandHandler {
			return app.PlaceOrderCommandHandler{Orders: port}
		},
		func(read a.PagedQueryPort[d.OrderState]) app.OrderingQueries { return app.OrderingQueries{Read: read} },
		orderingHTTPHandlers,
	), fx.Invoke(mountOrdering))
}

// OrderingHandlerDependencies declares the complete HTTP/request handler graph.
type OrderingHandlerDependencies struct {
	fx.In
	Queries        app.OrderingQueries
	CreateOrder    app.CreateOrderCommandHandler
	AddLine        app.AddLineCommandHandler
	ChangeQuantity app.ChangeQuantityCommandHandler
	PlaceOrder     app.PlaceOrderCommandHandler
}

func orderingHTTPHandlers(p OrderingHandlerDependencies) web.OrderingHTTPHandlers {
	return web.OrderingHTTPHandlers{OrderingQueries: p.Queries,
		CreateOrder:    p.CreateOrder,
		AddLine:        p.AddLine,
		ChangeQuantity: p.ChangeQuantity,
		PlaceOrder:     p.PlaceOrder,
	}
}
func mountOrdering(s *support.StorefrontRuntime, h web.OrderingHTTPHandlers, directory a.ProjectionPort[model.MenuPublished]) {
	web.Mount(s.Mux, h)
	s.RequestWorkers["ordering"] = rpc.Bind(rpc.OrderingRequestHandlers{Queries: h.OrderingQueries, CreateOrder: h.CreateOrder, AddLine: h.AddLine, ChangeQuantity: h.ChangeQuantity, PlaceOrder: h.PlaceOrder}).Run
	support.Subscribe(s, string(MenuDirectorySubscription), func(p model.MenuPublished) string { return p.EditionID }, app.MenuDirectoryProjectionHandler{Directory: directory}.Handle)
}

// OrderingSubscription identifies owner projection delivery.
type OrderingSubscription string

const MenuDirectorySubscription OrderingSubscription = "ordering.menu-directory"
