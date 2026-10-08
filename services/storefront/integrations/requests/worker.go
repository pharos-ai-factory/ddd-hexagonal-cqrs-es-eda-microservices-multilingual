package requests

import (
	"context"
	"fmt"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/menu"
	ordering "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/diagnostics"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"time"
)

func (r *Registry) Run(ctx context.Context, url string) {
	go r.runKind(ctx, url, "command")
	r.runKind(ctx, url, "query")
}
func (r *Registry) runKind(ctx context.Context, url, kind string) {
	for ctx.Err() == nil {
		err := broker.ServeRequests(ctx, url, r.Owner, kind, func(ctx context.Context, id string, data []byte) ([]byte, error) {
			var request RequestEnvelope
			if r.Owner == "menu" {
				request = &menu.Request{}
			} else if r.Owner == "ordering" {
				request = &ordering.Request{}
			} else {
				return nil, fmt.Errorf("request owner is unavailable")
			}
			if err := proto.Unmarshal(data, request); err != nil {
				return nil, err
			}
			if request.GetRequestId() != id {
				return nil, fmt.Errorf("request identity mismatch")
			}
			field, _, selectErr := Payload(request)
			if selectErr != nil || field.JSONName() != kind {
				return nil, fmt.Errorf("request kind disagrees with queue")
			}
			if _, _, err := Validate(request, r.Owner); err != nil {
				return nil, err
			}
			intent := &postgres.ReplyIntent{ID: id, Encode: func(outcome a.Outcome) ([]byte, error) {
				reply := OutcomeReply(r.Owner, outcome)
				reply.ProtoReflect().Set(reply.ProtoReflect().Descriptor().Fields().ByName("request_id"), protoreflect.ValueOfString(id))
				return proto.MarshalOptions{Deterministic: true}.Marshal(reply)
			}}
			if kind == "command" {
				ctx = postgres.WithReply(ctx, intent)
			}
			reply := r.Handle(ctx, request)
			if intent.Committed {
				return nil, nil
			}
			return proto.Marshal(reply)
		})
		if ctx.Err() != nil {
			return
		}
		diagnostics.Record(r.Owner, "requests.reconnect", "", "", err, false)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
