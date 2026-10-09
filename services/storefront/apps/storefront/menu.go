package main

import (
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/postgres"
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/queries"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	menuprojections "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/projections"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"go.uber.org/fx"
)

func menuModule() fx.Option {
	return fx.Module("menu", fx.Provide(
		func(s *support.StorefrontRuntime) execution.Transaction[*d.Drink] {
			return pg.NewDrinkTransaction(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) execution.Transaction[*d.MenuEdition] {
			return pg.NewEditionTransaction(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) ports.DrinkReadRepository {
			return pg.NewDrinkReadRepository(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) ports.EditionReadRepository {
			return pg.NewEditionReadRepository(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) a.ProjectionPort[app.DrinkPublished] {
			return store.Project[app.DrinkPublished](s.Databases["menu"], "published-drinks")
		},
		func(port execution.Transaction[*d.Drink]) execution.Executor[menucommands.CreateDrinkCommand] {
			return execution.Bind(port, func(repository ports.DrinkWriteRepository) menucommands.CreateDrinkCommandHandler {
				return menucommands.CreateDrinkCommandHandler{Repository: repository}
			})
		},
		func(port execution.Transaction[*d.Drink]) execution.Executor[menucommands.PublishDrinkCommand] {
			return execution.Bind(port, func(repository ports.DrinkWriteRepository) menucommands.PublishDrinkCommandHandler {
				return menucommands.PublishDrinkCommandHandler{Repository: repository}
			})
		},
		func(port execution.Transaction[*d.Drink]) execution.Executor[menucommands.ReviseDrinkCommand] {
			return execution.Bind(port, func(repository ports.DrinkWriteRepository) menucommands.ReviseDrinkCommandHandler {
				return menucommands.ReviseDrinkCommandHandler{Repository: repository}
			})
		},
		func(port execution.Transaction[*d.MenuEdition]) execution.Executor[menucommands.CreateEditionCommand] {
			return execution.Bind(port, func(repository ports.EditionWriteRepository) menucommands.CreateEditionCommandHandler {
				return menucommands.CreateEditionCommandHandler{Repository: repository}
			})
		},
		func(port execution.Transaction[*d.MenuEdition], directory a.ProjectionPort[app.DrinkPublished]) execution.Executor[menucommands.AddOfferCommand] {
			return execution.BindProjection(port, directory, func(c menucommands.AddOfferCommand) string { return fmt.Sprintf("%s/%d", c.DrinkID, c.DrinkRevision) },
				func(repository ports.EditionWriteRepository, directory a.ProjectionPort[app.DrinkPublished]) menucommands.AddOfferCommandHandler {
					return menucommands.AddOfferCommandHandler{Repository: repository, Drinks: directory}
				})
		},
		func(port execution.Transaction[*d.MenuEdition]) execution.Executor[menucommands.ChangePriceCommand] {
			return execution.Bind(port, func(repository ports.EditionWriteRepository) menucommands.ChangePriceCommandHandler {
				return menucommands.ChangePriceCommandHandler{Repository: repository}
			})
		},
		func(port execution.Transaction[*d.MenuEdition]) execution.Executor[menucommands.PublishEditionCommand] {
			return execution.Bind(port, func(repository ports.EditionWriteRepository) menucommands.PublishEditionCommandHandler {
				return menucommands.PublishEditionCommandHandler{Repository: repository}
			})
		},
		func(read ports.DrinkReadRepository) q.GetDrinkQueryHandler {
			return q.GetDrinkQueryHandler{ReadRepository: read}
		},
		func(read ports.DrinkReadRepository) q.ListDrinksQueryHandler {
			return q.ListDrinksQueryHandler{ReadRepository: read}
		},
		func(get q.GetDrinkQueryHandler, list q.ListDrinksQueryHandler) endpoints.DrinkQueryEndpoints {
			return endpoints.DrinkQueryEndpoints{GetHandler: get, ListHandler: list}
		},
		func(read ports.EditionReadRepository) q.GetEditionQueryHandler {
			return q.GetEditionQueryHandler{ReadRepository: read}
		},
		func(read ports.EditionReadRepository) q.ListEditionsQueryHandler {
			return q.ListEditionsQueryHandler{ReadRepository: read}
		},
		func(get q.GetEditionQueryHandler, list q.ListEditionsQueryHandler) endpoints.EditionQueryEndpoints {
			return endpoints.EditionQueryEndpoints{GetHandler: get, ListHandler: list}
		},
		menuRequestHandlers,
	), fx.Invoke(mountMenu))
}

// MenuHandlerDependencies belongs to composition; application types remain framework-free.
type MenuHandlerDependencies struct {
	fx.In
	DrinkQueries   endpoints.DrinkQueryEndpoints
	EditionQueries endpoints.EditionQueryEndpoints
	CreateDrink    execution.Executor[menucommands.CreateDrinkCommand]
	PublishDrink   execution.Executor[menucommands.PublishDrinkCommand]
	ReviseDrink    execution.Executor[menucommands.ReviseDrinkCommand]
	CreateEdition  execution.Executor[menucommands.CreateEditionCommand]
	AddOffer       execution.Executor[menucommands.AddOfferCommand]
	ChangePrice    execution.Executor[menucommands.ChangePriceCommand]
	PublishEdition execution.Executor[menucommands.PublishEditionCommand]
}

func menuRequestHandlers(p MenuHandlerDependencies) rpc.MenuRequestHandlers {
	return rpc.MenuRequestHandlers{DrinkQueries: p.DrinkQueries, EditionQueries: p.EditionQueries,
		CreateDrink:    p.CreateDrink,
		PublishDrink:   p.PublishDrink,
		ReviseDrink:    p.ReviseDrink,
		CreateEdition:  p.CreateEdition,
		AddOffer:       p.AddOffer,
		ChangePrice:    p.ChangePrice,
		PublishEdition: p.PublishEdition,
	}
}
func mountMenu(s *support.StorefrontRuntime, handlers rpc.MenuRequestHandlers, directory a.ProjectionPort[app.DrinkPublished]) {
	s.RequestWorkers["menu"] = rpc.Bind(handlers).Run
	support.SubscribePrivate(s, broker.Binding{Consumer: string(DrinkDirectorySubscription), Event: "menu.drink-published", Context: "menu", Visibility: "domain"}, func(p app.DrinkPublished) string { return p.DrinkID }, menuprojections.DrinkDirectoryProjectionHandler{Directory: directory}.Handle)
}

// MenuSubscription identifies private projection delivery without untyped routing literals.
type MenuSubscription string

const DrinkDirectorySubscription MenuSubscription = "menu.drink-directory"
