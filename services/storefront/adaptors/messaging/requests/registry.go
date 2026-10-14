package requests

import (
	"context"
	"fmt"
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
func Command[P proto.Message, I any](r *RabbitMQRequestRegistry, name string, convert func(P) I, handle func(context.Context, a.Metadata, I) (a.Outcome, error)) {
	r.bind(name, func(ctx context.Context, request RequestEnvelope, body proto.Message) (pb.Reply, error) {
		value, ok := body.(P)
		if !ok {
			return nil, fmt.Errorf("command payload type disagrees with its mapping")
		}
		input := convert(value)
		wire := CommandMetadata(request)
		metadata := a.Metadata{ID: wire.CommandId, AggregateID: wire.AggregateId, ExpectedVersion: wire.ExpectedVersion, CorrelationID: wire.CorrelationId, Name: r.Owner + "." + string(body.ProtoReflect().Descriptor().Name()), Input: input}
		result, err := handle(ctx, metadata, input)
		if err != nil {
			return nil, err
		}
		return OutcomeReply(r.Owner, result), nil
	})
}
func Queries[S any](r *RabbitMQRequestRegistry, singular, plural string, list func(context.Context) ([]a.Loaded[S], error), get func(context.Context, string) (a.Loaded[S], error), page func(context.Context, a.PageRequest) (a.Page[S], error), itemReply func(a.Loaded[S]) proto.Message, listReply func([]a.Loaded[S], bool, string) proto.Message) {
	upper := func(s string) string { return string(s[0]-32) + s[1:] }
	r.bind("get"+upper(singular), func(ctx context.Context, _ RequestEnvelope, body proto.Message) (pb.Reply, error) {
		input, ok := body.(interface{ GetId() string })
		if !ok {
			return nil, fmt.Errorf("item query has no identity")
		}
		loaded, err := get(ctx, input.GetId())
		if err != nil {
			return nil, err
		}
		if !loaded.Exists {
			reply, _ := pb.NewReply(r.Owner, "")
			return reply, SetPayload(reply, "error", &pb.RequestError{Code: "not_found"})
		}
		reply, _ := pb.NewReply(r.Owner, "")
		return reply, SetPayload(reply, singular, itemReply(loaded))
	})
	r.bind("list"+upper(plural), func(ctx context.Context, _ RequestEnvelope, body proto.Message) (pb.Reply, error) {
		input, ok := body.(interface{ GetPage() *pb.PageRequest })
		if !ok {
			return nil, fmt.Errorf("list query has no page input")
		}
		request := input.GetPage()
		var values []a.Loaded[S]
		var next string
		var err error
		if request != nil {
			if page == nil {
				return nil, fmt.Errorf("paged query unavailable")
			}
			var loaded a.Page[S]
			loaded, err = page(ctx, a.PageRequest{Limit: int(request.Limit), After: request.GetAfter()})
			values, next = loaded.Items, loaded.NextID
		} else {
			values, err = list(ctx)
		}
		if err != nil {
			return nil, err
		}
		reply, _ := pb.NewReply(r.Owner, "")
		return reply, SetPayload(reply, plural, listReply(values, request != nil, next))
	})
}

func OutcomeReply(owner string, result a.Outcome) pb.Reply {
	outcome := &pb.Outcome{AggregateId: result.AggregateID, Version: result.Version, Status: result.Status}
	if result.Rejection != nil {
		outcome.Rejection = &pb.Rejection{Code: result.Rejection.Code, Message: result.Rejection.Message}
	}
	reply, err := pb.NewReply(owner, "")
	if err != nil {
		panic(err)
	}
	if err = SetPayload(reply, "outcome", outcome); err != nil {
		panic(err)
	}
	return reply
}
