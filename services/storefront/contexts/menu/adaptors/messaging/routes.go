package messaging

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/integrations/requests"
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

func Bind(h Handlers) *rpc.Registry {
	r := rpc.New("menu")
	rpc.Command(r, "createDrink", func(p *pb.CreateDrink) app.CreateDrink { return app.CreateDrink{Name: p.GetName()} }, h.CreateDrink.Execute)
	rpc.Command(r, "publishDrink", func(p *pb.PublishDrink) struct{} { return struct{}{} }, h.PublishDrink.Execute)
	rpc.Command(r, "reviseDrink", func(p *pb.ReviseDrink) app.CreateDrink { return app.CreateDrink{Name: p.GetName()} }, h.ReviseDrink.Execute)
	rpc.Command(r, "createEdition", func(p *pb.CreateEdition) app.CreateEdition { return app.CreateEdition{Currency: p.GetCurrency()} }, h.CreateEdition.Execute)
	rpc.Command(r, "addOffer", func(p *pb.AddOffer) app.AddOffer {
		return app.AddOffer{Code: p.GetCode(), DrinkID: p.GetDrinkId(), DrinkRevision: p.GetDrinkRevision(), Minor: p.GetMinor()}
	}, h.AddOffer.Execute)
	rpc.Command(r, "changePrice", func(p *pb.ChangePrice) app.ChangePrice {
		return app.ChangePrice{Code: p.GetCode(), Minor: p.GetMinor()}
	}, h.ChangePrice.Execute)
	rpc.Command(r, "publishEdition", func(p *pb.PublishEdition) struct{} { return struct{}{} }, h.PublishEdition.Execute)
	rpc.Queries(r, "drink", "drinks", h.DrinkQueries.List, h.DrinkQueries.Get, h.DrinkQueries.Page, drinkReply, drinksReply)
	rpc.Queries(r, "edition", "editions", h.EditionQueries.List, h.EditionQueries.Get, h.EditionQueries.Page, editionReply, editionsReply)
	return r
}
