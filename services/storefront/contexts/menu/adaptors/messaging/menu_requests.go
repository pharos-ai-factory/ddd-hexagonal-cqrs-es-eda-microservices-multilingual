package messaging

import (
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/requests"
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/queryendpoints"
	commands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/commands"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
)

// CreateDrink maps the published payload to the application command.
func CreateDrink(p *pb.CreateDrink) commands.CreateDrinkCommand {
	return commands.CreateDrinkCommand{Name: p.GetName()}

}

// PublishDrink maps the published payload to the application command.
func PublishDrink(p *pb.PublishDrink) commands.PublishDrinkCommand {
	return commands.PublishDrinkCommand{}
}

// ReviseDrink maps the published payload to the application command.
func ReviseDrink(p *pb.ReviseDrink) commands.ReviseDrinkCommand {
	return commands.ReviseDrinkCommand{Name: p.GetName()}

}

// CreateEdition maps the published payload to the application command.
func CreateEdition(p *pb.CreateEdition) commands.CreateEditionCommand {
	return commands.CreateEditionCommand{Currency: p.GetCurrency()}

}

// AddOffer maps the published payload to the application command.
func AddOffer(p *pb.AddOffer) commands.AddOfferCommand {
	return commands.AddOfferCommand{Code: p.GetCode(), DrinkID: p.GetDrinkId(), DrinkRevision: p.GetDrinkRevision(), Minor: p.GetMinor()}

}

// ChangePrice maps the published payload to the application command.
func ChangePrice(p *pb.ChangePrice) commands.ChangePriceCommand {
	return commands.ChangePriceCommand{Code: p.GetCode(), Minor: p.GetMinor()}

}

// PublishEdition maps the published payload to the application command.
func PublishEdition(p *pb.PublishEdition) commands.PublishEditionCommand {
	return commands.PublishEditionCommand{}

}

// RegisterQueries mounts the context read endpoints.
func RegisterQueries(r *rpc.RabbitMQRequestRegistry, drinks endpoints.DrinkQueryEndpoints, editions endpoints.EditionQueryEndpoints) {
	rpc.Queries(r, "drink", "drinks", drinks.List, drinks.Get, drinks.Page, drinkReply, drinksReply)
	rpc.Queries(r, "edition", "editions", editions.List, editions.Get, editions.Page, editionReply, editionsReply)
}
