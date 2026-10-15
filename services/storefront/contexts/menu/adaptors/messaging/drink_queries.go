package messaging

import (
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetDrinkReply maps the resource revision or the existing missing-resource response.
func GetDrinkReply(_ menucontracts.GetDrinkQuery, loaded a.Loaded[view.DrinkView]) envelope.Reply {
	if !loaded.Exists {
		return &pb.Reply{Payload: &pb.Reply_Error{Error: &envelope.RequestError{Code: "not_found"}}}
	}
	return &pb.Reply{Payload: &pb.Reply_Drink{Drink: drinkReply(loaded).(*pb.LoadedDrink)}}
}

// ListDrinksReply retains pagination and its continuation from the handler result.
func ListDrinksReply(query menucontracts.ListDrinksQuery, page a.Page[view.DrinkView]) envelope.Reply {
	return &pb.Reply{Payload: &pb.Reply_Drinks{Drinks: drinksReply(page.Items, query.Page != nil, page.NextID).(*pb.Drinks)}}
}
