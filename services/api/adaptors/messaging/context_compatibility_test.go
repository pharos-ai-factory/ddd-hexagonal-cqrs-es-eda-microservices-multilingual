package messaging

import (
	"bytes"
	"encoding/hex"
	collection "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/collection"
	communication "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/communication"
	loyalty "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/loyalty"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/menu"
	ordering "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/ordering"
	preparation "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1/contexts/preparation"
	"google.golang.org/protobuf/proto"
	"os"
	"strings"
	"testing"
)

func TestContextRequestsPreserveHistoricalWireBytes(t *testing.T) {
	factories := map[string]func() proto.Message{"menu": func() proto.Message { return &menu.Request{} }, "ordering": func() proto.Message { return &ordering.Request{} }, "preparation": func() proto.Message { return &preparation.Request{} }, "collection": func() proto.Message { return &collection.Request{} }, "loyalty": func() proto.Message { return &loyalty.Request{} }, "communication": func() proto.Message { return &communication.Request{} }}
	for owner, factory := range factories {
		for _, kind := range []string{"query", "command"} {
			if owner == "communication" && kind == "command" {
				continue
			}
			t.Run(owner+"/"+kind, func(t *testing.T) {
				text, err := os.ReadFile("../../../../contracts/" + owner + "/messaging/fixtures/historical-" + kind + ".hex")
				if err != nil {
					t.Fatal(err)
				}
				body, err := hex.DecodeString(strings.TrimSpace(string(text)))
				if err != nil {
					t.Fatal(err)
				}
				request := factory()
				if err = proto.Unmarshal(body, request); err != nil {
					t.Fatal(err)
				}
				field, outer, err := Payload(request)
				if err != nil || field.JSONName() != kind {
					t.Fatal("historical payload was not recognised")
				}
				_, value, err := Payload(outer)
				if err != nil || !strings.Contains(string(value.ProtoReflect().Descriptor().FullName()), "cafe."+owner+".requests.v1.") {
					t.Fatal("payload does not belong to its owner")
				}
				encoded, err := proto.MarshalOptions{Deterministic: true}.Marshal(request)
				if err != nil || !bytes.Equal(encoded, body) {
					t.Fatal("historical wire bytes changed")
				}
			})
		}
	}
}
