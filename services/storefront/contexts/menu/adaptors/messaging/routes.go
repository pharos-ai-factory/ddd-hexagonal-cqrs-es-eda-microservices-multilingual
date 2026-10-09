package messaging

import (
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/queries"
	menucommands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/command"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/integrations/requests"
)

// MenuRequestHandlers groups the explicitly typed handlers mounted by this transport adaptor.
type MenuRequestHandlers struct {
	DrinkQueries   endpoints.DrinkQueryEndpoints
	CreateDrink    execution.Executor[menucommands.CreateDrinkCommand]
	PublishDrink   execution.Executor[menucommands.PublishDrinkCommand]
	ReviseDrink    execution.Executor[menucommands.ReviseDrinkCommand]
	EditionQueries endpoints.EditionQueryEndpoints
	CreateEdition  execution.Executor[menucommands.CreateEditionCommand]
	AddOffer       execution.Executor[menucommands.AddOfferCommand]
	ChangePrice    execution.Executor[menucommands.ChangePriceCommand]
	PublishEdition execution.Executor[menucommands.PublishEditionCommand]
}

func Bind(h MenuRequestHandlers) *rpc.RabbitMQRequestRegistry {
	r := rpc.New("menu")
	rpc.Command(r, "createDrink", func(p *pb.CreateDrink) menucommands.CreateDrinkCommand {
		return menucommands.CreateDrinkCommand{Name: p.GetName()}
	}, h.CreateDrink.Execute)
	rpc.Command(r, "publishDrink", func(p *pb.PublishDrink) menucommands.PublishDrinkCommand { return menucommands.PublishDrinkCommand{} }, h.PublishDrink.Execute)
	rpc.Command(r, "reviseDrink", func(p *pb.ReviseDrink) menucommands.ReviseDrinkCommand {
		return menucommands.ReviseDrinkCommand{Name: p.GetName()}
	}, h.ReviseDrink.Execute)
	rpc.Command(r, "createEdition", func(p *pb.CreateEdition) menucommands.CreateEditionCommand {
		return menucommands.CreateEditionCommand{Currency: p.GetCurrency()}
	}, h.CreateEdition.Execute)
	rpc.Command(r, "addOffer", func(p *pb.AddOffer) menucommands.AddOfferCommand {
		return menucommands.AddOfferCommand{Code: p.GetCode(), DrinkID: p.GetDrinkId(), DrinkRevision: p.GetDrinkRevision(), Minor: p.GetMinor()}
	}, h.AddOffer.Execute)
	rpc.Command(r, "changePrice", func(p *pb.ChangePrice) menucommands.ChangePriceCommand {
		return menucommands.ChangePriceCommand{Code: p.GetCode(), Minor: p.GetMinor()}
	}, h.ChangePrice.Execute)
	rpc.Command(r, "publishEdition", func(p *pb.PublishEdition) menucommands.PublishEditionCommand {
		return menucommands.PublishEditionCommand{}
	}, h.PublishEdition.Execute)
	rpc.Queries(r, "drink", "drinks", h.DrinkQueries.List, h.DrinkQueries.Get, h.DrinkQueries.Page, drinkReply, drinksReply)
	rpc.Queries(r, "edition", "editions", h.EditionQueries.List, h.EditionQueries.Get, h.EditionQueries.Page, editionReply, editionsReply)
	return r
}
