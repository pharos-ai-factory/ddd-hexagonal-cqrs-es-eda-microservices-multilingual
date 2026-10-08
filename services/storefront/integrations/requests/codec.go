// Package requests translates versioned wire messages at the transport boundary.
package requests

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var identity = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func ValidID(id string) bool { return identity.MatchString(id) }

func Payload(message proto.Message) (protoreflect.FieldDescriptor, proto.Message, error) {
	if message == nil {
		return nil, nil, fmt.Errorf("missing envelope")
	}
	value := message.ProtoReflect()
	field := value.WhichOneof(value.Descriptor().Oneofs().ByName("payload"))
	if field == nil {
		return nil, nil, fmt.Errorf("missing typed payload")
	}
	return field, value.Get(field).Message().Interface(), nil
}
func SetPayload(message proto.Message, name string, payload proto.Message) error {
	value := message.ProtoReflect()
	var field protoreflect.FieldDescriptor
	fields := value.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		if fields.Get(i).JSONName() == name {
			field = fields.Get(i)
			break
		}
	}
	if field == nil || field.Message() == nil || field.ContainingOneof() == nil {
		return fmt.Errorf("unknown payload: %s", name)
	}
	if payload == nil {
		payload = value.NewField(field).Message().Interface()
	}
	if field.Message().FullName() != payload.ProtoReflect().Descriptor().FullName() {
		return fmt.Errorf("wrong payload type")
	}
	value.Set(field, protoreflect.ValueOfMessage(payload.ProtoReflect()))
	return nil
}
func NewPayload(message proto.Message, name string) (proto.Message, error) {
	value := message.ProtoReflect()
	fields := value.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.JSONName() == name && field.Message() != nil {
			return value.NewField(field).Message().Interface(), nil
		}
	}
	return nil, fmt.Errorf("unknown payload: %s", name)
}

// Object preserves numeric HTTP/application values, explicit scalar defaults and
// optional-field absence. Protobuf JSON's quoted int64 representation stays here.
func Object(message proto.Message) map[string]any {
	value := message.ProtoReflect()
	fields := value.Descriptor().Fields()
	result := map[string]any{}
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.HasPresence() && !value.Has(field) {
			continue
		}
		v := value.Get(field)
		if field.IsList() {
			list := v.List()
			items := make([]any, 0, list.Len())
			for j := 0; j < list.Len(); j++ {
				items = append(items, scalar(field, list.Get(j)))
			}
			result[field.JSONName()] = items
		} else {
			result[field.JSONName()] = scalar(field, v)
		}
	}
	return result
}
func scalar(field protoreflect.FieldDescriptor, value protoreflect.Value) any {
	switch field.Kind() {
	case protoreflect.MessageKind:
		return Object(value.Message().Interface())
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Sint32Kind, protoreflect.Sint64Kind:
		return value.Int()
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		return value.Uint()
	default:
		return value.Interface()
	}
}
func FromObject(value any, message proto.Message) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return protojson.Unmarshal(data, message)
}
func Validate(request RequestEnvelope, owner string) (string, proto.Message, error) {
	if request.GetContractVersion() != 1 || request.GetContext() != owner || !ValidID(request.GetRequestId()) {
		return "", nil, fmt.Errorf("invalid request envelope")
	}
	kind, outer, err := Payload(request)
	if err != nil {
		return "", nil, err
	}
	field, body, err := Payload(outer)
	if err != nil {
		return "", nil, err
	}
	if !strings.Contains(field.Message().ParentFile().Path(), "/contexts/"+owner+"/") {
		return "", nil, fmt.Errorf("foreign context request")
	}
	if kind.JSONName() == "command" {
		metadata := CommandMetadata(request)
		if metadata == nil || !ValidID(metadata.CommandId) || !ValidID(metadata.AggregateId) || !ValidID(metadata.CorrelationId) || metadata.ExpectedVersion == nil {
			return "", nil, fmt.Errorf("invalid command metadata")
		}
		if err := RequiredInputs(body); err != nil {
			return "", nil, err
		}

	} else {
		object := Object(body)
		if id, present := object["id"]; present && !ValidID(id.(string)) {
			return "", nil, fmt.Errorf("invalid query identity")
		}
		if page, present := object["page"]; present {
			values := page.(map[string]any)
			limit := values["limit"].(uint64)
			if limit < 1 || limit > 100 {
				return "", nil, fmt.Errorf("invalid page limit")
			}
			if after, ok := values["after"]; ok && !ValidID(after.(string)) {
				return "", nil, fmt.Errorf("invalid page identity")
			}
		}
	}
	return field.JSONName(), body, nil
}

type RequestEnvelope interface {
	proto.Message
	GetContractVersion() uint32
	GetRequestId() string
	GetContext() string
}

func CommandMetadata(request RequestEnvelope) *pb.CommandMetadata {
	_, outer, err := Payload(request)
	if err != nil {
		return nil
	}
	value := outer.ProtoReflect()
	field := value.Descriptor().Fields().ByName("metadata")
	if field == nil || !value.Has(field) {
		return nil
	}
	metadata, _ := value.Get(field).Message().Interface().(*pb.CommandMetadata)
	return metadata
}

// RequiredInputs preserves scalar presence without treating optional additions as mandatory.
func RequiredInputs(message proto.Message) error {
	value := message.ProtoReflect()
	fields := value.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if proto.GetExtension(field.Options(), pb.E_RequiredInput).(bool) && !value.Has(field) {
			return fmt.Errorf("missing required input: %s", field.Name())
		}
	}
	return nil
}
