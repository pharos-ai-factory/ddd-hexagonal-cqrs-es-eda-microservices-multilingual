package messaging

import (
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/integrations/requests"
)

// MenuRequestHandlers groups the explicitly typed handlers mounted by this transport adaptor.
type MenuRequestHandlers struct {
	DrinkQueries   app.DrinkQueries
	CreateDrink    app.CreateDrinkCommandHandler
	PublishDrink   app.PublishDrinkCommandHandler
	ReviseDrink    app.ReviseDrinkCommandHandler
	EditionQueries app.EditionQueries
	CreateEdition  app.CreateEditionCommandHandler
	AddOffer       app.AddOfferCommandHandler
	ChangePrice    app.ChangePriceCommandHandler
	PublishEdition app.PublishEditionCommandHandler
}

func Bind(h MenuRequestHandlers) *rpc.RabbitMQRequestRegistry {
	r := rpc.New("menu")
	rpc.Command(r, "createDrink", func(p *pb.CreateDrink) app.CreateDrinkCommand { return app.CreateDrinkCommand{Name: p.GetName()} }, h.CreateDrink.Execute)
	rpc.Command(r, "publishDrink", func(p *pb.PublishDrink) app.PublishDrinkCommand { return app.PublishDrinkCommand{} }, h.PublishDrink.Execute)
	rpc.Command(r, "reviseDrink", func(p *pb.ReviseDrink) app.ReviseDrinkCommand { return app.ReviseDrinkCommand{Name: p.GetName()} }, h.ReviseDrink.Execute)
	rpc.Command(r, "createEdition", func(p *pb.CreateEdition) app.CreateEditionCommand {
		return app.CreateEditionCommand{Currency: p.GetCurrency()}
	}, h.CreateEdition.Execute)
	rpc.Command(r, "addOffer", func(p *pb.AddOffer) app.AddOfferCommand {
		return app.AddOfferCommand{Code: p.GetCode(), DrinkID: p.GetDrinkId(), DrinkRevision: p.GetDrinkRevision(), Minor: p.GetMinor()}
	}, h.AddOffer.Execute)
	rpc.Command(r, "changePrice", func(p *pb.ChangePrice) app.ChangePriceCommand {
		return app.ChangePriceCommand{Code: p.GetCode(), Minor: p.GetMinor()}
	}, h.ChangePrice.Execute)
	rpc.Command(r, "publishEdition", func(p *pb.PublishEdition) app.PublishEditionCommand { return app.PublishEditionCommand{} }, h.PublishEdition.Execute)
	rpc.Queries(r, "drink", "drinks", h.DrinkQueries.List, h.DrinkQueries.Get, h.DrinkQueries.Page, drinkReply, drinksReply)
	rpc.Queries(r, "edition", "editions", h.EditionQueries.List, h.EditionQueries.Get, h.EditionQueries.Page, editionReply, editionsReply)
	return r
}
