package requests

import (
	"context"
	"fmt"
	responses "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/generated/responses"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
)

type Handler func(context.Context, RequestEnvelope, proto.Message) (pb.Reply, error)

// RabbitMQRequestRegistry dispatches typed owner RabbitMQ requests to application handlers.
type RabbitMQRequestRegistry struct {
	Owner    string
	handlers map[string]Handler
}

func New(owner string) *RabbitMQRequestRegistry {
	if _, err := pb.NewReply(owner, ""); err != nil {
		panic(err)
	}
	return &RabbitMQRequestRegistry{Owner: owner, handlers: map[string]Handler{}}
}
func (r *RabbitMQRequestRegistry) bind(name string, handler Handler) {
	if r.handlers[name] != nil {
		panic("duplicate request handler")
	}
	r.handlers[name] = handler
}
func (r *RabbitMQRequestRegistry) Handle(ctx context.Context, request RequestEnvelope) pb.Reply {
	name, body, err := Validate(request, r.Owner)
	reply, _ := pb.NewReply(r.Owner, request.GetRequestId())
	if err != nil || r.handlers[name] == nil {
		_ = SetPayload(reply, "error", &pb.RequestError{Code: "invalid_request"})
		return reply
	}
	result, err := r.handlers[name](ctx, request, body)
	if err != nil {
		_ = SetPayload(reply, "error", &pb.RequestError{Code: "temporarily_unavailable"})
		return reply
	}
	field, body, err := Payload(result)
	if err != nil {
		panic(err)
	}
	if err = SetPayload(reply, field.JSONName(), body); err != nil {
		panic(err)
	}
	return reply
}
func Command[P proto.Message, I any](r *RabbitMQRequestRegistry, name pb.CommandName, convert func(P) I, handle func(context.Context, a.Metadata, I) (a.Outcome, error)) {
	if name == nil || name.Owner() != r.Owner || !name.Valid() {
		panic("command identifier disagrees with owner")
	}
	r.bind(name.String(), func(ctx context.Context, request RequestEnvelope, body proto.Message) (pb.Reply, error) {
		value, ok := body.(P)
		if !ok {
			return nil, fmt.Errorf("command payload type disagrees with its mapping")
		}
		input := convert(value)
		wire := CommandMetadata(request)
		metadata := a.Metadata{ID: wire.CommandId, AggregateID: wire.AggregateId, ExpectedVersion: wire.ExpectedVersion, CorrelationID: wire.CorrelationId, ActorID: wire.ActorId, Name: r.Owner + "." + string(body.ProtoReflect().Descriptor().Name()), Input: input}
		result, err := handle(ctx, metadata, input)
		if err != nil {
			return nil, err
		}
		return OutcomeReply(r.Owner, result)
	})
}

// Query connects one generated query identifier to its owner parser, handler and reply mapper.
func Query[P proto.Message, Q, R any](r *RabbitMQRequestRegistry, name pb.QueryName, parse func(P) Q, execute func(context.Context, Q) (R, error), reply func(Q, R) (pb.Reply, error)) {
	if name == nil || name.Owner() != r.Owner || !name.Valid() {
		panic("query identifier disagrees with owner")
	}
	var bus a.QueryBusPort[Q, R] = handlerQueryBus[Q, R]{execute}
	r.bind(name.String(), func(ctx context.Context, _ RequestEnvelope, body proto.Message) (pb.Reply, error) {
		value, ok := body.(P)
		if !ok {
			return nil, fmt.Errorf("query payload type disagrees with its mapping")
		}
		query := parse(value)
		result, err := bus.Ask(ctx, query)
		if err != nil {
			return nil, err
		}
		return reply(query, result)
	})
}

func OutcomeReply(owner string, result a.Outcome) (pb.Reply, error) {
	outcome, err := responses.ToOutcome(result)
	if err != nil {
		return nil, err
	}
	reply, err := pb.NewReply(owner, "")
	if err != nil {
		return nil, err
	}
	if err = SetPayload(reply, "outcome", outcome); err != nil {
		return nil, err
	}
	return reply, nil
}

// handlerQueryBus routes the application DTO after owner transport parsing.
type handlerQueryBus[Q, R any] struct {
	execute func(context.Context, Q) (R, error)
}

func (b handlerQueryBus[Q, R]) Ask(ctx context.Context, query Q) (R, error) {
	return b.execute(ctx, query)
}
