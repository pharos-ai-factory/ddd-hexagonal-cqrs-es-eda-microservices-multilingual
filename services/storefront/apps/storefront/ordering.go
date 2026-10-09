package main

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/http"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/postgres"
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/queries"
	orderingcommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	orderingprojections "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/projections"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/queries"
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
		func(s *support.StorefrontRuntime) ports.OrderReader {
			return pg.OrderReader(s.Databases["ordering"])
		},
		func(s *support.StorefrontRuntime) a.ProjectionPort[model.MenuPublished] {
			return store.Project[model.MenuPublished](s.Databases["ordering"], "published-menus")
		},
		func(port a.AggregateCommandPort[d.OrderState], directory a.ProjectionPort[model.MenuPublished]) orderingcommands.CreateOrderCommandHandler {
			return orderingcommands.CreateOrderCommandHandler{Orders: port, Menus: directory}
		},
		func(port a.AggregateCommandPort[d.OrderState], directory a.ProjectionPort[model.MenuPublished]) orderingcommands.AddLineCommandHandler {
			return orderingcommands.AddLineCommandHandler{Orders: port, Menus: directory}
		},
		func(port a.AggregateCommandPort[d.OrderState]) orderingcommands.ChangeQuantityCommandHandler {
			return orderingcommands.ChangeQuantityCommandHandler{Orders: port}
		},
		func(port a.AggregateCommandPort[d.OrderState]) orderingcommands.PlaceOrderCommandHandler {
			return orderingcommands.PlaceOrderCommandHandler{Orders: port}
		},
		func(read ports.OrderReader) q.GetOrderQueryHandler { return q.GetOrderQueryHandler{Read: read} },
		func(read ports.OrderReader) q.ListOrdersQueryHandler { return q.ListOrdersQueryHandler{Read: read} },
		func(get q.GetOrderQueryHandler, list q.ListOrdersQueryHandler) endpoints.OrderQueryEndpoints {
			return endpoints.OrderQueryEndpoints{GetHandler: get, ListHandler: list}
		},
		orderingHTTPHandlers,
	), fx.Invoke(mountOrdering))
}

// OrderingHandlerDependencies declares the complete HTTP/request handler graph.
type OrderingHandlerDependencies struct {
	fx.In
	Queries        endpoints.OrderQueryEndpoints
	CreateOrder    orderingcommands.CreateOrderCommandHandler
	AddLine        orderingcommands.AddLineCommandHandler
	ChangeQuantity orderingcommands.ChangeQuantityCommandHandler
	PlaceOrder     orderingcommands.PlaceOrderCommandHandler
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
	support.Subscribe(s, string(MenuDirectorySubscription), func(p model.MenuPublished) string { return p.EditionID }, orderingprojections.MenuDirectoryProjectionHandler{Directory: directory}.Handle)
}

// OrderingSubscription identifies owner projection delivery.
type OrderingSubscription string

const MenuDirectorySubscription OrderingSubscription = "ordering.menu-directory"
