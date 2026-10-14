package messaging

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
)

func drinkReply(value a.Loaded[d.DrinkView]) proto.Message {
	return &pb.LoadedDrink{Exists: value.Exists, Version: value.Version, State: &pb.Drink{Id: value.State.ID, Name: value.State.Name, Revision: value.State.Revision, Published: value.State.Published}}
}
func drinksReply(values []a.Loaded[d.DrinkView], paged bool, next string) proto.Message {
	result := &pb.Drinks{Items: make([]*pb.LoadedDrink, 0, len(values)), Paged: paged}
	for _, item := range values {
		result.Items = append(result.Items, drinkReply(item).(*pb.LoadedDrink))
	}
	if next != "" {
		result.NextId = &next
	}
	return result
}
