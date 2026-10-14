package di

import (
	"fmt"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	requests "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/requests"
	registration "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/commandregistration"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging"
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/queryendpoints"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

type menuRequests struct {
	*requests.RabbitMQRequestRegistry
}

func menuRequestsProvider(drinkTransaction execution.Transaction[*d.Drink], editionTransaction execution.Transaction[*d.MenuEdition], directory a.ProjectionPort[app.DrinkPublishedDomainEvent], drinks endpoints.DrinkQueryEndpoints, editions endpoints.EditionQueryEndpoints) menuRequests {
	r := requests.New("menu")
	drinkCommands := registration.New(r, drinkTransaction)
	editionCommands := registration.New(r, editionTransaction)
	registration.Register(drinkCommands, "createDrink", rpc.CreateDrink, func(repository ports.DrinkWriteRepository) menucommands.CreateDrinkCommandHandler {
		return menucommands.CreateDrinkCommandHandler{Repository: repository}
	})
	registration.Register(drinkCommands, "publishDrink", rpc.PublishDrink, func(repository ports.DrinkWriteRepository) menucommands.PublishDrinkCommandHandler {
		return menucommands.PublishDrinkCommandHandler{Repository: repository}
	})
	registration.Register(drinkCommands, "reviseDrink", rpc.ReviseDrink, func(repository ports.DrinkWriteRepository) menucommands.ReviseDrinkCommandHandler {
		return menucommands.ReviseDrinkCommandHandler{Repository: repository}
	})
	registration.Register(editionCommands, "createEdition", rpc.CreateEdition, func(repository ports.EditionWriteRepository) menucommands.CreateEditionCommandHandler {
		return menucommands.CreateEditionCommandHandler{Repository: repository}
	})
	registration.RegisterProjection(editionCommands, "addOffer", rpc.AddOffer, directory, func(c menucommands.AddOfferCommand) string { return fmt.Sprintf("%s/%d", c.DrinkID, c.DrinkRevision) },
		func(repository ports.EditionWriteRepository, directory a.ProjectionPort[app.DrinkPublishedDomainEvent]) menucommands.AddOfferCommandHandler {
			return menucommands.AddOfferCommandHandler{Repository: repository, Drinks: directory}
		})
	registration.Register(editionCommands, "changePrice", rpc.ChangePrice, func(repository ports.EditionWriteRepository) menucommands.ChangePriceCommandHandler {
		return menucommands.ChangePriceCommandHandler{Repository: repository}
	})
	registration.Register(editionCommands, "publishEdition", rpc.PublishEdition, func(repository ports.EditionWriteRepository) menucommands.PublishEditionCommandHandler {
		return menucommands.PublishEditionCommandHandler{Repository: repository}
	})
	rpc.RegisterQueries(r, drinks, editions)
	return menuRequests{r}
}
