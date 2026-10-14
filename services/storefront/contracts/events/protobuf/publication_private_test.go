package protobuf

import (
	"bytes"
	"encoding/hex"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	order "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"os"
	"testing"
)

func TestPrivatePublicationFixturesPreserveEvidence(t *testing.T) {
	for _, path := range []string{"menu/adaptors/messaging/fixtures/edition-published.domain.hex", "ordering/adaptors/messaging/fixtures/order-placed.domain.hex"} {
		t.Run(path, func(t *testing.T) {
			text, err := os.ReadFile("../../../contexts/" + path)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := hex.DecodeString(string(bytes.TrimSpace(text)))
			if err != nil {
				t.Fatal(err)
			}
			event, err := Decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := Encode(event)
			if err != nil || !bytes.Equal(encoded, raw) {
				t.Fatalf("private bytes changed: %v", err)
			}
			original := event
			event.Context = "communication"
			if _, err := Encode(event); err == nil {
				t.Fatal("foreign owner accepted")
			}
			event = original
			switch fact := event.Payload.(type) {
			case menu.MenuEditionPublishedDomainEvent:
				fact.State.Offers[0].Currency = "GBP"
				event.Payload = fact
			case order.OrderPlacedDomainEvent:
				fact.State.Lines[0].Quantity = 6
				event.Payload = fact
			default:
				t.Fatal("untyped private domain event")
			}
			if _, err := Encode(event); err == nil {
				t.Fatal("invalid immutable evidence accepted")
			}
		})
	}
}
