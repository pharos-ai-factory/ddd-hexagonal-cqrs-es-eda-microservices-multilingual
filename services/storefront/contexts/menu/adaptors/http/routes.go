package http

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	web "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/http"
	"net/http"
)

type Handlers struct {
	DrinkQueries   app.DrinkQueries
	CreateDrink    app.CreateDrinkHandler
	PublishDrink   app.PublishDrinkHandler
	ReviseDrink    app.ReviseDrinkHandler
	EditionQueries app.EditionQueries
	CreateEdition  app.CreateEditionHandler
	AddOffer       app.AddOfferHandler
	ChangePrice    app.ChangePriceHandler
	PublishEdition app.PublishEditionHandler
}

func Mount(mux *http.ServeMux, h Handlers) {
	mux.HandleFunc("GET /v1/menu/drinks", web.List(h.DrinkQueries.List))
	mux.HandleFunc("GET /v1/menu/drinks/{id}", web.Get(h.DrinkQueries.Get))
	mux.HandleFunc("POST /v1/menu/drinks/{id}", web.Command("menu.CreateDrink", h.CreateDrink.Execute))
	mux.HandleFunc("POST /v1/menu/drinks/{id}/publish", web.Command("menu.PublishDrink", h.PublishDrink.Execute))
	mux.HandleFunc("POST /v1/menu/drinks/{id}/revise", web.Command("menu.ReviseDrink", h.ReviseDrink.Execute))
	mux.HandleFunc("GET /v1/menu/editions", web.List(h.EditionQueries.List))
	mux.HandleFunc("GET /v1/menu/editions/{id}", web.Get(h.EditionQueries.Get))
	mux.HandleFunc("POST /v1/menu/editions/{id}", web.Command("menu.CreateEdition", h.CreateEdition.Execute))
	mux.HandleFunc("POST /v1/menu/editions/{id}/offers", web.Command("menu.AddOffer", h.AddOffer.Execute, "minor"))
	mux.HandleFunc("POST /v1/menu/editions/{id}/prices", web.Command("menu.ChangePrice", h.ChangePrice.Execute, "minor"))
	mux.HandleFunc("POST /v1/menu/editions/{id}/publish", web.Command("menu.PublishEdition", h.PublishEdition.Execute))
}
