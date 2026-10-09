package http

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	"net/http"
)

// MenuHTTPHandlers groups the explicitly typed handlers mounted by this transport adaptor.
type MenuHTTPHandlers struct {
	DrinkQueries   app.DrinkQueries
	CreateDrink    menucommands.CreateDrinkCommandHandler
	PublishDrink   menucommands.PublishDrinkCommandHandler
	ReviseDrink    menucommands.ReviseDrinkCommandHandler
	EditionQueries app.EditionQueries
	CreateEdition  menucommands.CreateEditionCommandHandler
	AddOffer       menucommands.AddOfferCommandHandler
	ChangePrice    menucommands.ChangePriceCommandHandler
	PublishEdition menucommands.PublishEditionCommandHandler
}

func Mount(mux interface {
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}, h MenuHTTPHandlers) {
	mux.HandleFunc("GET /v1/menu/drinks", web.PagedList(h.DrinkQueries.List, h.DrinkQueries.Page))
	mux.HandleFunc("GET /v1/menu/drinks/{id}", web.Get(h.DrinkQueries.Get))
	mux.HandleFunc("POST /v1/menu/drinks/{id}", web.Command("menu.CreateDrink", h.CreateDrink.Execute))
	mux.HandleFunc("POST /v1/menu/drinks/{id}/publish", web.Command("menu.PublishDrink", h.PublishDrink.Execute))
	mux.HandleFunc("POST /v1/menu/drinks/{id}/revise", web.Command("menu.ReviseDrink", h.ReviseDrink.Execute))
	mux.HandleFunc("GET /v1/menu/editions", web.PagedList(h.EditionQueries.List, h.EditionQueries.Page))
	mux.HandleFunc("GET /v1/menu/editions/{id}", web.Get(h.EditionQueries.Get))
	mux.HandleFunc("POST /v1/menu/editions/{id}", web.Command("menu.CreateEdition", h.CreateEdition.Execute))
	mux.HandleFunc("POST /v1/menu/editions/{id}/offers", web.Command("menu.AddOffer", h.AddOffer.Execute, "minor"))
	mux.HandleFunc("POST /v1/menu/editions/{id}/prices", web.Command("menu.ChangePrice", h.ChangePrice.Execute, "minor"))
	mux.HandleFunc("POST /v1/menu/editions/{id}/publish", web.Command("menu.PublishEdition", h.PublishEdition.Execute))
}
