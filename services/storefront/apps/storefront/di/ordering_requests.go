package di

import (
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	requests "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/requests"
	registration "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/commandregistration"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging/queryendpoints"
	orderingcommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type orderingRequests struct {
	*requests.RabbitMQRequestRegistry
}

func orderingRequestsProvider(transaction execution.Transaction[*d.Order], directory a.ProjectionPort[model.MenuPublishedIntegrationEvent], orders endpoints.OrderQueryEndpoints) orderingRequests {
	r := requests.New("ordering")
	commands := registration.New(r, transaction)
	registration.RegisterProjection(commands, "createOrder", rpc.CreateOrder, directory, func(c orderingcommands.CreateOrderCommand) string { return c.EditionID }, func(repository ports.OrderWriteRepository, directory a.ProjectionPort[model.MenuPublishedIntegrationEvent]) orderingcommands.CreateOrderCommandHandler {
		return orderingcommands.CreateOrderCommandHandler{Repository: repository, Menus: directory}
	})
	registration.RegisterProjection(commands, "addLine", rpc.AddLine, directory, func(c orderingcommands.AddLineCommand) string { return c.EditionID }, func(repository ports.OrderWriteRepository, directory a.ProjectionPort[model.MenuPublishedIntegrationEvent]) orderingcommands.AddLineCommandHandler {
		return orderingcommands.AddLineCommandHandler{Repository: repository, Menus: directory}
	})
	registration.Register(commands, "changeQuantity", rpc.ChangeQuantity, func(repository ports.OrderWriteRepository) orderingcommands.ChangeQuantityCommandHandler {
		return orderingcommands.ChangeQuantityCommandHandler{Repository: repository}
	})
	registration.Register(commands, "placeOrder", rpc.PlaceOrder, func(repository ports.OrderWriteRepository) orderingcommands.PlaceOrderCommandHandler {
		return orderingcommands.PlaceOrderCommandHandler{Repository: repository}
	})
	rpc.RegisterQueries(r, orders)
	return orderingRequests{r}
}
