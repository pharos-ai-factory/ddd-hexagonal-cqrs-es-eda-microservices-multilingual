package messaging

import (
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	menucontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetEditionReply maps the resource revision or the existing missing-resource response.
func GetEditionReply(_ menucontracts.GetEditionQuery, loaded a.Loaded[view.EditionView]) envelope.Reply {
	if !loaded.Exists {
		return &pb.Reply{Payload: &pb.Reply_Error{Error: &envelope.RequestError{Code: "not_found"}}}
	}
	return &pb.Reply{Payload: &pb.Reply_Edition{Edition: editionReply(loaded).(*pb.LoadedEdition)}}
}

// ListEditionsReply retains pagination and its continuation from the handler result.
func ListEditionsReply(query menucontracts.ListEditionsQuery, page a.Page[view.EditionView]) envelope.Reply {
	return &pb.Reply{Payload: &pb.Reply_Editions{Editions: editionsReply(page.Items, query.Page != nil, page.NextID).(*pb.Editions)}}
}
