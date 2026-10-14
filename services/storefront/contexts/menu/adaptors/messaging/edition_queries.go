package messaging

import (
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/readmodels"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetEdition decodes the published item query into its application input.
func GetEdition(value *pb.GetEdition) q.GetEditionQuery { return q.GetEditionQuery{ID: value.GetId()} }

// ListEditions preserves complete reads and explicitly requested pagination.
func ListEditions(value *pb.ListEditions) q.ListEditionsQuery {
	result := q.ListEditionsQuery{}
	if page := value.GetPage(); page != nil {
		result.Page = &a.PageRequest{Limit: int(page.Limit), After: page.GetAfter()}
	}
	return result
}

// GetEditionReply maps the resource revision or the existing missing-resource response.
func GetEditionReply(_ q.GetEditionQuery, loaded a.Loaded[view.EditionView]) envelope.Reply {
	if !loaded.Exists {
		return &pb.Reply{Payload: &pb.Reply_Error{Error: &envelope.RequestError{Code: "not_found"}}}
	}
	return &pb.Reply{Payload: &pb.Reply_Edition{Edition: editionReply(loaded).(*pb.LoadedEdition)}}
}

// ListEditionsReply retains pagination and its continuation from the handler result.
func ListEditionsReply(query q.ListEditionsQuery, page a.Page[view.EditionView]) envelope.Reply {
	return &pb.Reply{Payload: &pb.Reply_Editions{Editions: editionsReply(page.Items, query.Page != nil, page.NextID).(*pb.Editions)}}
}
