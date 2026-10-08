package messaging

import (
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"testing"
)

func TestFutureOptionalInputDoesNotBecomeMandatory(t *testing.T) {
	required := &descriptorpb.FieldOptions{}
	proto.SetExtension(required, pb.E_RequiredInput, true)
	descriptor := &descriptorpb.FileDescriptorProto{Name: proto.String("future-command.proto"), Package: proto.String("future"), Syntax: proto.String("proto3"), Dependency: []string{"cafe/requests/v1/validation.proto"}, MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Command"), OneofDecl: []*descriptorpb.OneofDescriptorProto{{Name: proto.String("_price")}, {Name: proto.String("_note")}}, Field: []*descriptorpb.FieldDescriptorProto{
		{Name: proto.String("price"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Proto3Optional: proto.Bool(true), OneofIndex: proto.Int32(0), Options: required},
		{Name: proto.String("note"), Number: proto.Int32(2), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Proto3Optional: proto.Bool(true), OneofIndex: proto.Int32(1)},
	}}}}
	file, err := protodesc.NewFile(descriptor, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	value := dynamicpb.NewMessage(file.Messages().Get(0))
	if RequiredInputs(value) == nil {
		t.Fatal("required price absence accepted")
	}
	value.Set(value.Descriptor().Fields().ByName("price"), protoreflect.ValueOfInt64(0))
	if err := RequiredInputs(value); err != nil {
		t.Fatal("zero price or omitted optional note rejected", err)
	}
	value.Set(value.Descriptor().Fields().ByName("note"), protoreflect.ValueOfString("future additive input"))
	if err := RequiredInputs(value); err != nil {
		t.Fatal(err)
	}
}
