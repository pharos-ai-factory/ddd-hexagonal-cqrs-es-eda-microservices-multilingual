package messaging

import (
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetDrink decodes the published item query into its application input.
func GetDrink(value *pb.GetDrink) q.GetDrinkQuery { return q.GetDrinkQuery{ID: value.GetId()} }

// ListDrinks preserves complete reads and explicitly requested pagination.
func ListDrinks(value *pb.ListDrinks) q.ListDrinksQuery {
	result := q.ListDrinksQuery{}
	if page := value.GetPage(); page != nil {
		result.Page = &a.PageRequest{Limit: int(page.Limit), After: page.GetAfter()}
	}
	return result
}

// GetDrinkReply maps the resource revision or the existing missing-resource response.
func GetDrinkReply(_ q.GetDrinkQuery, loaded a.Loaded[view.DrinkView]) envelope.Reply {
	if !loaded.Exists {
		return &pb.Reply{Payload: &pb.Reply_Error{Error: &envelope.RequestError{Code: "not_found"}}}
	}
	return &pb.Reply{Payload: &pb.Reply_Drink{Drink: drinkReply(loaded).(*pb.LoadedDrink)}}
}

// ListDrinksReply retains pagination and its continuation from the handler result.
func ListDrinksReply(query q.ListDrinksQuery, page a.Page[view.DrinkView]) envelope.Reply {
	return &pb.Reply{Payload: &pb.Reply_Drinks{Drinks: drinksReply(page.Items, query.Page != nil, page.NextID).(*pb.Drinks)}}
}
