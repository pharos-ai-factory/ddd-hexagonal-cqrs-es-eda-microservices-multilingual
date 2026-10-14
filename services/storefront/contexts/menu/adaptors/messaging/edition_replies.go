package messaging

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
)

func editionReply(value a.Loaded[d.EditionView]) proto.Message {
	state := &pb.Edition{Id: value.State.ID, Currency: value.State.Currency, Status: value.State.Status, Offers: make([]*pb.Offer, 0, len(value.State.Offers))}
	for _, offer := range value.State.Offers {
		state.Offers = append(state.Offers, &pb.Offer{Code: offer.Code, DrinkId: offer.DrinkID, DrinkRevision: offer.DrinkRevision, Name: offer.Name, Minor: offer.Minor, Currency: offer.Currency})
	}
	return &pb.LoadedEdition{Exists: value.Exists, Version: value.Version, State: state}
}
func editionsReply(values []a.Loaded[d.EditionView], paged bool, next string) proto.Message {
	result := &pb.Editions{Items: make([]*pb.LoadedEdition, 0, len(values)), Paged: paged}
	for _, item := range values {
		result.Items = append(result.Items, editionReply(item).(*pb.LoadedEdition))
	}
	if next != "" {
		result.NextId = &next
	}
	return result
}
