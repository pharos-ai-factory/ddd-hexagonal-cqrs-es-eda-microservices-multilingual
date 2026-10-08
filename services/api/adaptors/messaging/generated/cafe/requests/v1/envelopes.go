// Code generated from context ownership. DO NOT EDIT.
package requestsv1

import (
	"fmt"
	collection "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/collection"
	communication "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/communication"
	loyalty "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/loyalty"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/menu"
	ordering "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/ordering"
	preparation "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/preparation"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/shared"
	"google.golang.org/protobuf/proto"
)

type Request interface {
	proto.Message
	GetContractVersion() uint32
	GetRequestId() string
	GetContext() string
}
type Reply interface {
	Request
	GetOutcome() *shared.Outcome
	GetError() *shared.RequestError
}
type CommandEnvelope interface {
	proto.Message
	GetMetadata() *shared.CommandMetadata
}
type PageRequest = shared.PageRequest
type CommandMetadata = shared.CommandMetadata
type Rejection = shared.Rejection
type Outcome = shared.Outcome
type RequestError = shared.RequestError

var E_RequiredInput = shared.E_RequiredInput

func NewRequest(owner, id string) (Request, error) {
	switch owner {
	case "menu":
		return &menu.Request{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "ordering":
		return &ordering.Request{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "preparation":
		return &preparation.Request{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "collection":
		return &collection.Request{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "loyalty":
		return &loyalty.Request{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "communication":
		return &communication.Request{ContractVersion: 1, RequestId: id, Context: owner}, nil
	}
	return nil, fmt.Errorf("unknown request owner: %s", owner)
}
func NewReply(owner, id string) (Reply, error) {
	switch owner {
	case "menu":
		return &menu.Reply{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "ordering":
		return &ordering.Reply{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "preparation":
		return &preparation.Reply{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "collection":
		return &collection.Reply{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "loyalty":
		return &loyalty.Reply{ContractVersion: 1, RequestId: id, Context: owner}, nil
	case "communication":
		return &communication.Reply{ContractVersion: 1, RequestId: id, Context: owner}, nil
	}
	return nil, fmt.Errorf("unknown request owner: %s", owner)
}
func Command(request Request) CommandEnvelope {
	switch value := request.(type) {
	case *menu.Request:
		if value.GetCommand() != nil {
			return value.GetCommand()
		}
	case *ordering.Request:
		if value.GetCommand() != nil {
			return value.GetCommand()
		}
	case *preparation.Request:
		if value.GetCommand() != nil {
			return value.GetCommand()
		}
	case *collection.Request:
		if value.GetCommand() != nil {
			return value.GetCommand()
		}
	case *loyalty.Request:
		if value.GetCommand() != nil {
			return value.GetCommand()
		}
	}
	return nil
}
func Query(request Request) proto.Message {
	switch value := request.(type) {
	case *menu.Request:
		if value.GetQuery() != nil {
			return value.GetQuery()
		}
	case *ordering.Request:
		if value.GetQuery() != nil {
			return value.GetQuery()
		}
	case *preparation.Request:
		if value.GetQuery() != nil {
			return value.GetQuery()
		}
	case *collection.Request:
		if value.GetQuery() != nil {
			return value.GetQuery()
		}
	case *loyalty.Request:
		if value.GetQuery() != nil {
			return value.GetQuery()
		}
	case *communication.Request:
		if value.GetQuery() != nil {
			return value.GetQuery()
		}
	}
	return nil
}
