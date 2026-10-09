package main

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
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
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	"go.uber.org/fx"
)

func orderingModule() fx.Option {
	return fx.Module("ordering", fx.Provide(
		func(s *support.StorefrontRuntime) execution.Transaction[*d.Order] {
			return pg.NewOrderTransaction(s.Databases["ordering"])
		},
		func(s *support.StorefrontRuntime) ports.OrderReadRepository {
			return pg.NewOrderReadRepository(s.Databases["ordering"])
		},
		func(s *support.StorefrontRuntime) a.ProjectionPort[model.MenuPublished] {
			return store.Project[model.MenuPublished](s.Databases["ordering"], "published-menus")
		},
		func(port execution.Transaction[*d.Order], directory a.ProjectionPort[model.MenuPublished]) execution.Executor[orderingcommands.CreateOrderCommand] {
			return execution.BindProjection(port, directory, func(c orderingcommands.CreateOrderCommand) string { return c.EditionID }, func(repository ports.OrderWriteRepository, directory a.ProjectionPort[model.MenuPublished]) orderingcommands.CreateOrderCommandHandler {
				return orderingcommands.CreateOrderCommandHandler{Repository: repository, Menus: directory}
			})
		},
		func(port execution.Transaction[*d.Order], directory a.ProjectionPort[model.MenuPublished]) execution.Executor[orderingcommands.AddLineCommand] {
			return execution.BindProjection(port, directory, func(c orderingcommands.AddLineCommand) string { return c.EditionID }, func(repository ports.OrderWriteRepository, directory a.ProjectionPort[model.MenuPublished]) orderingcommands.AddLineCommandHandler {
				return orderingcommands.AddLineCommandHandler{Repository: repository, Menus: directory}
			})
		},
		func(port execution.Transaction[*d.Order]) execution.Executor[orderingcommands.ChangeQuantityCommand] {
			return execution.Bind(port, func(repository ports.OrderWriteRepository) orderingcommands.ChangeQuantityCommandHandler {
				return orderingcommands.ChangeQuantityCommandHandler{Repository: repository}
			})
		},
		func(port execution.Transaction[*d.Order]) execution.Executor[orderingcommands.PlaceOrderCommand] {
			return execution.Bind(port, func(repository ports.OrderWriteRepository) orderingcommands.PlaceOrderCommandHandler {
				return orderingcommands.PlaceOrderCommandHandler{Repository: repository}
			})
		},
		func(read ports.OrderReadRepository) q.GetOrderQueryHandler {
			return q.GetOrderQueryHandler{ReadRepository: read}
		},
		func(read ports.OrderReadRepository) q.ListOrdersQueryHandler {
			return q.ListOrdersQueryHandler{ReadRepository: read}
		},
		func(get q.GetOrderQueryHandler, list q.ListOrdersQueryHandler) endpoints.OrderQueryEndpoints {
			return endpoints.OrderQueryEndpoints{GetHandler: get, ListHandler: list}
		},
		orderingRequestHandlers,
	), fx.Invoke(mountOrdering))
}

// OrderingHandlerDependencies declares the complete RabbitMQ request handler graph.
type OrderingHandlerDependencies struct {
	fx.In
	Queries        endpoints.OrderQueryEndpoints
	CreateOrder    execution.Executor[orderingcommands.CreateOrderCommand]
	AddLine        execution.Executor[orderingcommands.AddLineCommand]
	ChangeQuantity execution.Executor[orderingcommands.ChangeQuantityCommand]
	PlaceOrder     execution.Executor[orderingcommands.PlaceOrderCommand]
}

func orderingRequestHandlers(p OrderingHandlerDependencies) rpc.OrderingRequestHandlers {
	return rpc.OrderingRequestHandlers{Queries: p.Queries,
		CreateOrder:    p.CreateOrder,
		AddLine:        p.AddLine,
		ChangeQuantity: p.ChangeQuantity,
		PlaceOrder:     p.PlaceOrder,
	}
}
func mountOrdering(s *support.StorefrontRuntime, h rpc.OrderingRequestHandlers, directory a.ProjectionPort[model.MenuPublished]) {
	s.RequestWorkers["ordering"] = rpc.Bind(h).Run
	support.Subscribe(s, string(MenuDirectorySubscription), func(p model.MenuPublished) string { return p.EditionID }, orderingprojections.MenuDirectoryProjectionHandler{Directory: directory}.Handle)
}

// OrderingSubscription identifies owner projection delivery.
type OrderingSubscription string

const MenuDirectorySubscription OrderingSubscription = "ordering.menu-directory"
