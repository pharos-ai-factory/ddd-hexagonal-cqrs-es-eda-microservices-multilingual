package di

import (
	"fmt"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	requests "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/requests"
	registration "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/commandregistration"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type menuRequests struct {
	*requests.RabbitMQRequestRegistry
}

func menuRequestsProvider(drinkTransaction execution.Transaction[*d.Drink], editionTransaction execution.Transaction[*d.MenuEdition], directory a.ProjectionPort[app.DrinkPublishedDomainEvent], getDrink q.GetDrinkQueryHandler, listDrinks q.ListDrinksQueryHandler, getEdition q.GetEditionQueryHandler, listEditions q.ListEditionsQueryHandler) menuRequests {
	r := requests.New("menu")
	drinkCommands := registration.New(r, drinkTransaction)
	editionCommands := registration.New(r, editionTransaction)
	registration.Register(drinkCommands, names.MenuCommandCreateDrink, rpc.CreateDrink, func(repository ports.DrinkWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) menucommands.CreateDrinkCommandHandler {
		return menucommands.CreateDrinkCommandHandler{Repository: repository, Realtime: realtime}
	})
	registration.Register(drinkCommands, names.MenuCommandPublishDrink, rpc.PublishDrink, func(repository ports.DrinkWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) menucommands.PublishDrinkCommandHandler {
		return menucommands.PublishDrinkCommandHandler{Repository: repository, Realtime: realtime}
	})
	registration.Register(drinkCommands, names.MenuCommandReviseDrink, rpc.ReviseDrink, func(repository ports.DrinkWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) menucommands.ReviseDrinkCommandHandler {
		return menucommands.ReviseDrinkCommandHandler{Repository: repository, Realtime: realtime}
	})
	registration.Register(editionCommands, names.MenuCommandCreateEdition, rpc.CreateEdition, func(repository ports.EditionWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) menucommands.CreateEditionCommandHandler {
		return menucommands.CreateEditionCommandHandler{Repository: repository, Realtime: realtime}
	})
	registration.RegisterProjection(editionCommands, names.MenuCommandAddOffer, rpc.AddOffer, directory, func(c menucontracts.AddOfferCommand) string { return fmt.Sprintf("%s/%d", c.DrinkID, c.DrinkRevision) },
		func(repository ports.EditionWriteRepository, directory a.ProjectionPort[app.DrinkPublishedDomainEvent], realtime a.RealtimeBusPort[a.RealtimeEvent]) menucommands.AddOfferCommandHandler {
			return menucommands.AddOfferCommandHandler{Repository: repository, Realtime: realtime, Drinks: directory}
		})
	registration.Register(editionCommands, names.MenuCommandChangePrice, rpc.ChangePrice, func(repository ports.EditionWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) menucommands.ChangePriceCommandHandler {
		return menucommands.ChangePriceCommandHandler{Repository: repository, Realtime: realtime}
	})
	registration.Register(editionCommands, names.MenuCommandPublishEdition, rpc.PublishEdition, func(repository ports.EditionWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) menucommands.PublishEditionCommandHandler {
		return menucommands.PublishEditionCommandHandler{Repository: repository, Realtime: realtime}
	})
	requests.Query(r, names.MenuQueryGetDrink, rpc.GetDrink, getDrink.Execute, rpc.GetDrinkReply)
	requests.Query(r, names.MenuQueryListDrinks, rpc.ListDrinks, listDrinks.Execute, rpc.ListDrinksReply)
	requests.Query(r, names.MenuQueryGetEdition, rpc.GetEdition, getEdition.Execute, rpc.GetEditionReply)
	requests.Query(r, names.MenuQueryListEditions, rpc.ListEditions, listEditions.Execute, rpc.ListEditionsReply)
	return menuRequests{r}
}
