package main

import (
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	menuhttp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/http"
	menupg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/postgres"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	orderhttp "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/http"
	orderpg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/postgres"
	ordering "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
)

func main() { support.Main(build) }
func build() (*support.Service, error) {
	s, err := support.Open("menu", "ordering")
	if err != nil {
		return nil, err
	}
	drinks := menupg.DrinkCommands(s.Databases["menu"])
	editions := menupg.EditionCommands(s.Databases["menu"])
	drinkDirectory := store.Project[model.DrinkPublished](s.Databases["menu"], "published-drinks")
	menus := store.Project[model.MenuPublished](s.Databases["ordering"], "published-menus")
	orders := orderpg.OrderCommands(s.Databases["ordering"])
	menuhttp.Mount(s.Mux, menuhttp.Handlers{
		DrinkQueries: menu.DrinkQueries{Read: menupg.DrinkQueries(s.Databases["menu"])}, EditionQueries: menu.EditionQueries{Read: menupg.EditionQueries(s.Databases["menu"])},
		CreateDrink: menu.CreateDrinkHandler{Drinks: drinks}, PublishDrink: menu.PublishDrinkHandler{Drinks: drinks}, ReviseDrink: menu.ReviseDrinkHandler{Drinks: drinks},
		CreateEdition: menu.CreateEditionHandler{Editions: editions}, AddOffer: menu.AddOfferHandler{Editions: editions, Drinks: drinkDirectory}, ChangePrice: menu.ChangePriceHandler{Editions: editions}, PublishEdition: menu.PublishEditionHandler{Editions: editions},
	})
	orderhttp.Mount(s.Mux, orderhttp.Handlers{OrderingQueries: ordering.OrderingQueries{Read: orderpg.OrderQueries(s.Databases["ordering"])}, CreateOrder: ordering.CreateOrderHandler{Orders: orders, Menus: menus}, AddLine: ordering.AddLineHandler{Orders: orders, Menus: menus}, ChangeQuantity: ordering.ChangeQuantityHandler{Orders: orders}, PlaceOrder: ordering.PlaceOrderHandler{Orders: orders}})
	support.Subscribe(s, "menu.drink-directory", func(p model.DrinkPublished) string { return p.DrinkID }, menu.ProjectDrinkHandler{Directory: drinkDirectory}.Handle)
	support.Subscribe(s, "ordering.menu-directory", func(p model.MenuPublished) string { return p.EditionID }, ordering.ProjectMenuHandler{Directory: menus}.Handle)
	return s, nil
}
