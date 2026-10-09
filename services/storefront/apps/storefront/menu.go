package main

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/http"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/postgres"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"go.uber.org/fx"
)

func menuModule() fx.Option {
	return fx.Module("menu", fx.Provide(
		func(s *support.StorefrontRuntime) a.AggregateCommandPort[d.DrinkState] {
			return pg.DrinkCommands(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) a.AggregateCommandPort[d.EditionState] {
			return pg.EditionCommands(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) a.PagedQueryPort[d.DrinkState] {
			return pg.DrinkQueries(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) a.PagedQueryPort[d.EditionState] {
			return pg.EditionQueries(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) a.ProjectionPort[app.DrinkPublished] {
			return store.Project[app.DrinkPublished](s.Databases["menu"], "published-drinks")
		},
		func(port a.AggregateCommandPort[d.DrinkState]) menucommands.CreateDrinkCommandHandler {
			return menucommands.CreateDrinkCommandHandler{Drinks: port}
		},
		func(port a.AggregateCommandPort[d.DrinkState]) menucommands.PublishDrinkCommandHandler {
			return menucommands.PublishDrinkCommandHandler{Drinks: port}
		},
		func(port a.AggregateCommandPort[d.DrinkState]) menucommands.ReviseDrinkCommandHandler {
			return menucommands.ReviseDrinkCommandHandler{Drinks: port}
		},
		func(port a.AggregateCommandPort[d.EditionState]) menucommands.CreateEditionCommandHandler {
			return menucommands.CreateEditionCommandHandler{Editions: port}
		},
		func(port a.AggregateCommandPort[d.EditionState], directory a.ProjectionPort[app.DrinkPublished]) menucommands.AddOfferCommandHandler {
			return menucommands.AddOfferCommandHandler{Editions: port, Drinks: directory}
		},
		func(port a.AggregateCommandPort[d.EditionState]) menucommands.ChangePriceCommandHandler {
			return menucommands.ChangePriceCommandHandler{Editions: port}
		},
		func(port a.AggregateCommandPort[d.EditionState]) menucommands.PublishEditionCommandHandler {
			return menucommands.PublishEditionCommandHandler{Editions: port}
		},
		func(read a.PagedQueryPort[d.DrinkState]) app.DrinkQueries { return app.DrinkQueries{Read: read} },
		func(read a.PagedQueryPort[d.EditionState]) app.EditionQueries { return app.EditionQueries{Read: read} },
		menuHTTPHandlers,
	), fx.Invoke(mountMenu))
}

// MenuHandlerDependencies belongs to composition; application types remain framework-free.
type MenuHandlerDependencies struct {
	fx.In
	DrinkQueries   app.DrinkQueries
	EditionQueries app.EditionQueries
	CreateDrink    menucommands.CreateDrinkCommandHandler
	PublishDrink   menucommands.PublishDrinkCommandHandler
	ReviseDrink    menucommands.ReviseDrinkCommandHandler
	CreateEdition  menucommands.CreateEditionCommandHandler
	AddOffer       menucommands.AddOfferCommandHandler
	ChangePrice    menucommands.ChangePriceCommandHandler
	PublishEdition menucommands.PublishEditionCommandHandler
}

func menuHTTPHandlers(p MenuHandlerDependencies) web.MenuHTTPHandlers {
	return web.MenuHTTPHandlers{DrinkQueries: p.DrinkQueries, EditionQueries: p.EditionQueries,
		CreateDrink:    p.CreateDrink,
		PublishDrink:   p.PublishDrink,
		ReviseDrink:    p.ReviseDrink,
		CreateEdition:  p.CreateEdition,
		AddOffer:       p.AddOffer,
		ChangePrice:    p.ChangePrice,
		PublishEdition: p.PublishEdition,
	}
}
func mountMenu(s *support.StorefrontRuntime, handlers web.MenuHTTPHandlers, directory a.ProjectionPort[app.DrinkPublished]) {
	web.Mount(s.Mux, handlers)
	s.RequestWorkers["menu"] = rpc.Bind(rpc.MenuRequestHandlers(handlers)).Run
	support.SubscribePrivate(s, broker.Binding{Consumer: string(DrinkDirectorySubscription), Event: "menu.drink-published", Context: "menu", Visibility: "domain"}, func(p app.DrinkPublished) string { return p.DrinkID }, app.DrinkDirectoryProjectionHandler{Directory: directory}.Handle)
}

// MenuSubscription identifies private projection delivery without untyped routing literals.
type MenuSubscription string

const DrinkDirectorySubscription MenuSubscription = "menu.drink-directory"
