package main

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	menuhttp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/http"
	menurpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging"
	menupg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/postgres"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	orderhttp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/http"
	orderrpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	orderpg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/postgres"
	ordering "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
)

func main() { support.Main(build) }
func build() (*support.Service, error) {
	s, err := support.Open("menu", "ordering")
	if err != nil {
		return nil, err
	}
	drinks := menupg.DrinkCommands(s.Databases["menu"])
	editions := menupg.EditionCommands(s.Databases["menu"])
	drinkDirectory := store.Project[menu.DrinkPublished](s.Databases["menu"], "published-drinks")
	menus := store.Project[model.MenuPublished](s.Databases["ordering"], "published-menus")
	orders := orderpg.OrderCommands(s.Databases["ordering"])
	menuHandlers := menuhttp.Handlers{
		DrinkQueries: menu.DrinkQueries{Read: menupg.DrinkQueries(s.Databases["menu"])}, EditionQueries: menu.EditionQueries{Read: menupg.EditionQueries(s.Databases["menu"])},
		CreateDrink: menu.CreateDrinkHandler{Drinks: drinks}, PublishDrink: menu.PublishDrinkHandler{Drinks: drinks}, ReviseDrink: menu.ReviseDrinkHandler{Drinks: drinks},
		CreateEdition: menu.CreateEditionHandler{Editions: editions}, AddOffer: menu.AddOfferHandler{Editions: editions, Drinks: drinkDirectory}, ChangePrice: menu.ChangePriceHandler{Editions: editions}, PublishEdition: menu.PublishEditionHandler{Editions: editions},
	}
	menuhttp.Mount(s.Mux, menuHandlers)
	s.RequestWorkers["menu"] = menurpc.Bind(menurpc.Handlers(menuHandlers)).Run
	orderHandlers := orderhttp.Handlers{OrderingQueries: ordering.OrderingQueries{Read: orderpg.OrderQueries(s.Databases["ordering"])}, CreateOrder: ordering.CreateOrderHandler{Orders: orders, Menus: menus}, AddLine: ordering.AddLineHandler{Orders: orders, Menus: menus}, ChangeQuantity: ordering.ChangeQuantityHandler{Orders: orders}, PlaceOrder: ordering.PlaceOrderHandler{Orders: orders}}
	orderhttp.Mount(s.Mux, orderHandlers)
	s.RequestWorkers["ordering"] = orderrpc.Bind(orderrpc.Handlers{Queries: orderHandlers.OrderingQueries, CreateOrder: orderHandlers.CreateOrder, AddLine: orderHandlers.AddLine, ChangeQuantity: orderHandlers.ChangeQuantity, PlaceOrder: orderHandlers.PlaceOrder}).Run
	support.SubscribePrivate(s, broker.Binding{Consumer: "menu.drink-directory", Event: "menu.drink-published", Context: "menu", Visibility: "domain"}, func(p menu.DrinkPublished) string { return p.DrinkID }, menu.ProjectDrinkHandler{Directory: drinkDirectory}.Handle)
	support.Subscribe(s, "ordering.menu-directory", func(p model.MenuPublished) string { return p.EditionID }, ordering.ProjectMenuHandler{Directory: menus}.Handle)
	return s, nil
}
