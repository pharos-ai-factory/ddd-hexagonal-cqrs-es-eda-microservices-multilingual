package protobuf

import (
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/cafe/v1"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"reflect"
	"strings"
	"testing"
	"time"
)

func fixture() a.Message {
	return a.Message{ID: "11111111-1111-4111-8111-111111111111", Name: "menu.edition-published", Context: "menu", Visibility: a.Public, ContractVersion: 1, AggregateKind: "edition", AggregateID: "22222222-2222-4222-8222-222222222222", AggregateVersion: 3, CorrelationID: "33333333-3333-4333-8333-333333333333", CausationID: "44444444-4444-4444-8444-444444444444", OccurredAt: time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC), Payload: model.MenuPublished{EditionID: "22222222-2222-4222-8222-222222222222", Currency: "EUR", Offers: []model.Offer{{Code: "TEA", DrinkID: "55555555-5555-4555-8555-555555555555", DrinkRevision: 1, Name: "Tea", Minor: 300, Currency: "EUR"}}}}
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
	payload := new(pb.MenuPublished)
	if err = proto.Unmarshal(event.Payload, payload); err != nil {
		t.Fatal(err)
	}
	payload.Offers[0].Minor = -1
	event.Payload, err = proto.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	body, err = proto.Marshal(&event)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(body); err == nil {
		t.Fatal("valid Protobuf with invalid business evidence was accepted")
	}
}
func TestDeclaredIntegrationContractRoundTrip(t *testing.T) {
	message := fixture()
	encoded, err := Encode(message)
	if err != nil {
		t.Fatal(err)
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
	message.Payload = model.OrderPlaced{OrderID: message.ID}
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

func TestGeneratedPayloadMappingRejectsIntegerOverflowBeforeConversion(t *testing.T) {
	message := fixture()
	message.Name, message.Context, message.AggregateKind = "ordering.order-placed", "ordering", "order"
	message.Payload = model.OrderPlaced{OrderID: message.AggregateID, CustomerID: message.ID, EditionID: message.ID, Currency: "EUR", Lines: []model.Line{{ID: message.ID, OfferCode: "TEA", Name: "Tea", Quantity: 4294967297, Minor: 0}}}
	if _, err := Encode(message); err == nil {
		t.Fatal("quantity overflow wrapped into an accepted value")
	}
}

func TestPublicEnvelopeRejectsUnsupportedRevisionAndContract(t *testing.T) {
	for _, change := range []func(*a.Message){
		func(m *a.Message) { m.AggregateVersion = 9007199254740992 },
		func(m *a.Message) { m.Name = "menu.unknown" },
	} {
		message := fixture()
		change(&message)
		if _, err := Encode(message); err == nil {
			t.Fatal("unsupported envelope accepted")
		}
	}
}
