// Package protobuf handles the technical envelope around generated event payload bindings.
package protobuf

import (
	"fmt"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/envelope"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/cafe/v1"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"time"
)

type payloadCodec struct {
	owner, kind string
	version     uint32
	encode      func(any) (proto.Message, error)
	decode      func(proto.Message) any
	newMessage  func() proto.Message
}

func Encode(message a.Message) ([]byte, error) {
	if message.Visibility == a.Private {
		binding, ok := privateCodecs[message.Context]
		if !ok {
			return nil, fmt.Errorf("unknown private owner")
		}
		return binding.encode(message)
	}
	if err := validateEnvelope(message); err != nil {
		return nil, err
	}
	binding := payloadCodecs[message.Name]
	payload, err := binding.encode(message.Payload)
	if err != nil {
		return nil, err
	}
	if err = validatePayload(payload.ProtoReflect(), message.AggregateID); err != nil {
		return nil, err
	}
	body, err := proto.MarshalOptions{Deterministic: true}.Marshal(payload)
	if err != nil {
		return nil, err
	}
	event := &pb.Event{Id: message.ID, Name: message.Name, Context: message.Context, Visibility: string(message.Visibility), ContractVersion: message.ContractVersion, AggregateKind: message.AggregateKind, AggregateId: message.AggregateID, AggregateVersion: message.AggregateVersion, CorrelationId: message.CorrelationID, CausationId: message.CausationID, OccurredAt: message.OccurredAt.UTC().Format(time.RFC3339Nano), Payload: body}
	encoded, err := proto.MarshalOptions{Deterministic: true}.Marshal(event)
	if len(encoded) > 256*1024 {
		return nil, fmt.Errorf("event exceeds size limit")
	}
	return encoded, err
}

func Decode(data []byte) (a.Message, error) {
	if len(data) > 256*1024 {
		return a.Message{}, fmt.Errorf("event exceeds size limit")
	}
	var event pb.Event
	if err := proto.Unmarshal(data, &event); err != nil {
		return a.Message{}, err
	}
	if event.Visibility == string(a.Private) {
		binding, ok := privateCodecs[event.Context]
		if !ok {
			return a.Message{}, fmt.Errorf("unknown private owner")
		}
		return binding.decode(data)
	}
	occurred, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil {
		return a.Message{}, err
	}
	message := a.Message{ID: event.Id, Name: event.Name, Context: event.Context, Visibility: a.Visibility(event.Visibility), ContractVersion: event.ContractVersion, AggregateKind: event.AggregateKind, AggregateID: event.AggregateId, AggregateVersion: event.AggregateVersion, CorrelationID: event.CorrelationId, CausationID: event.CausationId, OccurredAt: occurred}
	if err = validateEnvelope(message); err != nil {
		return a.Message{}, err
	}
	if len(event.Payload) == 0 {
		return a.Message{}, fmt.Errorf("missing payload")
	}
	binding := payloadCodecs[message.Name]
	payload := binding.newMessage()
	if err = proto.Unmarshal(event.Payload, payload); err != nil {
		return a.Message{}, err
	}
	if err = validatePayload(payload.ProtoReflect(), message.AggregateID); err != nil {
		return a.Message{}, err
	}
	message.Payload = binding.decode(payload)
	return message, nil
}

func validateEnvelope(m a.Message) error {
	definition, ok := payloadCodecs[m.Name]
	if !ok || definition.owner != m.Context || m.Visibility != a.Public || m.AggregateKind != definition.kind || m.ContractVersion != definition.version {
		return fmt.Errorf("unknown event contract or invalid owner/visibility")
	}
	return envelope.ValidateMetadata(m)
}

type privateCodec struct {
	encode func(a.Message) ([]byte, error)
	decode func([]byte) (a.Message, error)
}
