package di

import (
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	requests "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/requests"
	registration "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/commandregistration"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	orderingcommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/queries"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type orderingRequests struct {
	*requests.RabbitMQRequestRegistry
	Commands registration.Registry[*d.Order]
}

func orderingRequestsProvider(transaction execution.Transaction[*d.Order], directory a.ProjectionPort[model.MenuPublishedIntegrationEvent], getOrder q.GetOrderQueryHandler, listOrders q.ListOrdersQueryHandler) orderingRequests {
	r := requests.New("ordering")
	commands := registration.New(r, transaction)
	registration.RegisterProjection(commands, names.OrderingCommandCreateOrder, rpc.CreateOrder, directory, func(c orderingcontracts.CreateOrderCommand) string { return c.EditionID }, func(repository ports.OrderWriteRepository, directory a.ProjectionPort[model.MenuPublishedIntegrationEvent], realtime a.RealtimeBusPort[a.RealtimeEvent]) orderingcommands.CreateOrderCommandHandler {
		return orderingcommands.CreateOrderCommandHandler{Repository: repository, Realtime: realtime, Menus: directory}
	})
	registration.RegisterProjection(commands, names.OrderingCommandAddLine, rpc.AddLine, directory, func(c orderingcontracts.AddLineCommand) string { return c.EditionID }, func(repository ports.OrderWriteRepository, directory a.ProjectionPort[model.MenuPublishedIntegrationEvent], realtime a.RealtimeBusPort[a.RealtimeEvent]) orderingcommands.AddLineCommandHandler {
		return orderingcommands.AddLineCommandHandler{Repository: repository, Realtime: realtime, Menus: directory}
	})
	registration.Register(commands, names.OrderingCommandChangeQuantity, rpc.ChangeQuantity, func(repository ports.OrderWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) orderingcommands.ChangeQuantityCommandHandler {
		return orderingcommands.ChangeQuantityCommandHandler{Repository: repository, Realtime: realtime}
	})
	registration.Register(commands, names.OrderingCommandPlaceOrder, rpc.PlaceOrder, func(repository ports.OrderWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) orderingcommands.PlaceOrderCommandHandler {
		return orderingcommands.PlaceOrderCommandHandler{Repository: repository}
	})
	requests.Query(r, names.OrderingQueryGetOrder, rpc.GetOrder, getOrder.Execute, rpc.GetOrderReply)
	requests.Query(r, names.OrderingQueryListOrders, rpc.ListOrders, listOrders.Execute, rpc.ListOrdersReply)
	return orderingRequests{r, commands}
}
