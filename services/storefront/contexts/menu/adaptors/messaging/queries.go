package messaging

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
)

func drinkReply(value a.Loaded[d.DrinkState]) proto.Message {
	return &pb.LoadedDrink{Exists: value.Exists, Version: value.Version, State: &pb.Drink{Id: value.State.ID, Name: value.State.Name, Revision: value.State.Revision, Published: value.State.Published}}
}
func drinksReply(values []a.Loaded[d.DrinkState], paged bool, next string) proto.Message {
	result := &pb.Drinks{Items: make([]*pb.LoadedDrink, 0, len(values)), Paged: paged}
	for _, item := range values {
		result.Items = append(result.Items, drinkReply(item).(*pb.LoadedDrink))
	}
	if next != "" {
		result.NextId = &next
	}
	return result
}
func editionReply(value a.Loaded[d.EditionState]) proto.Message {
	state := &pb.Edition{Id: value.State.ID, Currency: value.State.Currency, Status: value.State.Status, Offers: make([]*pb.Offer, 0, len(value.State.Offers))}
	for _, offer := range value.State.Offers {
		state.Offers = append(state.Offers, &pb.Offer{Code: offer.Code, DrinkId: offer.DrinkID, DrinkRevision: offer.DrinkRevision, Name: offer.Name, Minor: offer.Minor, Currency: offer.Currency})
	}
	return &pb.LoadedEdition{Exists: value.Exists, Version: value.Version, State: state}
}
func editionsReply(values []a.Loaded[d.EditionState], paged bool, next string) proto.Message {
	result := &pb.Editions{Items: make([]*pb.LoadedEdition, 0, len(values)), Paged: paged}
	for _, item := range values {
		result.Items = append(result.Items, editionReply(item).(*pb.LoadedEdition))
	}
	if next != "" {
		result.NextId = &next
	}
	return result
}
