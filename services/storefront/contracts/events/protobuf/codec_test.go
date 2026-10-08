package protobuf

import (
	"bytes"
	"encoding/hex"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/cafe/v1"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func fixture() a.Message {
	return a.Message{ID: "11111111-1111-4111-8111-111111111111", Name: "loyalty.reward-issued", Context: "loyalty", Visibility: a.Public, ContractVersion: 1, AggregateKind: "reward", AggregateID: "22222222-2222-4222-8222-222222222222", AggregateVersion: 3, CorrelationID: "33333333-3333-4333-8333-333333333333", CausationID: "44444444-4444-4444-8444-444444444444", OccurredAt: time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC), Payload: model.RewardIssued{RewardID: "22222222-2222-4222-8222-222222222222", CustomerID: "55555555-5555-4555-8555-555555555555", Benefit: "one free drink", ExpiresAt: "2026-10-03T12:00:00Z"}}
}

func TestDecoderRejectsSemanticallyInvalidWirePayload(t *testing.T) {
	body, err := Encode(fixture())
	if err != nil {
		t.Fatal(err)
	}
	var event pb.Event
	if err = proto.Unmarshal(body, &event); err != nil {
		t.Fatal(err)
	}
	event.GetRewardIssued().ExpiresAt = "invalid"
	body, err = proto.Marshal(&event)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(body); err == nil {
		t.Fatal("valid Protobuf with invalid business evidence was accepted")
	}
}
func TestGoldenIntegrationContract(t *testing.T) {
	message := fixture()
	encoded, err := Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("..", "fixtures", "reward-issued.v1.hex")
	if os.Getenv("UPDATE_FIXTURES") == "1" {
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, []byte(hex.EncodeToString(encoded)+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	golden, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytes.TrimSpace(golden)) != hex.EncodeToString(encoded) {
		t.Fatal("wire fixture changed; review the versioned contract")
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(message, decoded) {
		t.Fatalf("round trip changed message: %+v", decoded)
	}
}
func TestClosedContractRejectsWrongScopeAndType(t *testing.T) {
	message := fixture()
	message.Visibility = a.Private
	if _, err := Encode(message); err == nil {
		t.Fatal("integration event acquired private scope")
	}
	message = fixture()
	message.Context = "communication"
	if _, err := Encode(message); err == nil {
		t.Fatal("foreign producer was accepted")
	}
	message = fixture()
	message.ContractVersion = 2
	if _, err := Encode(message); err == nil {
		t.Fatal("unknown version accepted")
	}
	message = fixture()
	message.Payload = model.OrderCollected{OrderID: message.ID}
	if _, err := Encode(message); err == nil {
		t.Fatal("payload/name mismatch accepted")
	}
	if _, err := Decode([]byte("not protobuf")); err == nil {
		t.Fatal("invalid wire data accepted")
	}
}

func TestEnvelopeRequiresCausationInEveryLanguage(t *testing.T) {
	t.Run("encode", func(t *testing.T) {
		message := fixture()
		message.CausationID = ""
		if _, err := Encode(message); err == nil {
			t.Fatal("a publication without its producing command identity was accepted")
		}
	})
	t.Run("decode", func(t *testing.T) {
		body, err := Encode(fixture())
		if err != nil {
			t.Fatal(err)
		}
		var message pb.Event
		if err = proto.Unmarshal(body, &message); err != nil {
			t.Fatal(err)
		}
		message.CausationId = ""
		body, err = proto.Marshal(&message)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = Decode(body); err == nil {
			t.Fatal("Go accepted missing causation that Python and TypeScript reject")
		}
	})
}

func TestPublishedDrinkPreservesAnEightyCharacterUnicodeName(t *testing.T) {
	message := fixture()
	message.Name, message.Context, message.AggregateKind, message.Visibility = "menu.drink-published", "menu", "drink", a.Private
	message.Payload = menu.DrinkPublished{DrinkID: message.AggregateID, Name: strings.Repeat("ή", 80), Revision: 1}
	body, err := Encode(message)
	if err != nil {
		t.Fatalf("the published contract rejected a name inside its character limit: %v", err)
	}
	if _, err = Decode(body); err != nil {
		t.Fatal(err)
	}
}
