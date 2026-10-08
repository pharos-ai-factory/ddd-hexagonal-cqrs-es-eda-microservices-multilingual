package protobuf

import (
	"bytes"
	"encoding/hex"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"os"
	"reflect"
	"testing"
)

func TestOwnerPrivateDeliveryPreservesHistoricalBytes(t *testing.T) {
	message := fixture()
	message.Name, message.Context, message.AggregateKind, message.Visibility = "menu.drink-published", "menu", "drink", a.Private
	message.Payload = menu.DrinkPublished{DrinkID: message.AggregateID, Name: "Coffee", Revision: 1}
	encoded, err := Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile("../../../contexts/menu/adaptors/messaging/fixtures/drink-published.hex")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := hex.DecodeString(string(bytes.TrimSpace(text)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, previous) {
		t.Fatal("private source relocation changed stored delivery bytes")
	}
	decoded, err := Decode(previous)
	if err != nil || !reflect.DeepEqual(decoded, message) {
		t.Fatal("stored owner-private message did not recover", err)
	}
}
