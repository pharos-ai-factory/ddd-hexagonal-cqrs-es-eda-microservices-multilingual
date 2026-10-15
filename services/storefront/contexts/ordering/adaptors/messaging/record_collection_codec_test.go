package messaging

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging/generated"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"os"
	"strings"
	"testing"
)

func collectionMetadata() a.Metadata {
	source := "11111111-1111-4111-8111-111111111111"
	consumer := RecordCollectionCodec{}.Consumer()
	return a.Metadata{ID: a.DerivedID(consumer, source), AggregateID: "22222222-2222-4222-8222-222222222222", Name: consumer, Consumer: consumer, SourceEventID: source, CausationID: source, CorrelationID: "33333333-3333-4333-8333-333333333333", SourceHash: strings.Repeat("a", 64), Input: map[string]any{"orderId": "22222222-2222-4222-8222-222222222222", "customerId": "33333333-3333-4333-8333-333333333333"}}
}
func TestRecordCollectionStoredBytesAndReceiptMaterial(t *testing.T) {
	codec := RecordCollectionCodec{}
	m := collectionMetadata()
	c := orderingcontracts.RecordCollectionCommand{CustomerID: m.CorrelationID}
	body, err := codec.Encode(m, c)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("fixtures/record_collection.command.hex")
	if err != nil {
		t.Fatal(err)
	}
	historical, err := hex.DecodeString(strings.TrimSpace(string(fixture)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, historical) {
		t.Fatal("historical private command bytes changed")
	}
	decoded, command, err := codec.Decode(historical)
	if err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(m.Input)
	receipt, _ := json.Marshal(decoded.Input)
	if decoded.ID != m.ID || decoded.SourceHash != m.SourceHash || decoded.CausationID != m.SourceEventID || !bytes.Equal(input, receipt) || command != c {
		t.Fatal("command identity or source receipt changed")
	}
	replay, err := codec.Encode(decoded, command)
	if err != nil || !bytes.Equal(body, replay) {
		t.Fatal("round-trip changed stored bytes")
	}
}
func TestRecordCollectionRejectsInvalidEnvelopes(t *testing.T) {
	codec := RecordCollectionCodec{}
	m := collectionMetadata()
	body, _ := codec.Encode(m, orderingcontracts.RecordCollectionCommand{CustomerID: m.CorrelationID})
	cases := map[string]func(*pb.CommandEnvelope){
		"foreign consumer": func(e *pb.CommandEnvelope) { e.Consumer = "loyalty.record-collection" },
		"changed identity": func(e *pb.CommandEnvelope) { e.Id = e.Target },
		"invalid target":   func(e *pb.CommandEnvelope) { e.Target = "invalid" },
		"invalid hash":     func(e *pb.CommandEnvelope) { e.SourceHash = "invalid" },
		"receipt array":    func(e *pb.CommandEnvelope) { e.ReceiptMaterial = []byte(`[]`) },
		"receipt null":     func(e *pb.CommandEnvelope) { e.ReceiptMaterial = []byte(`null`) },
		"missing command":  func(e *pb.CommandEnvelope) { e.Command = nil },
		"invalid customer": func(e *pb.CommandEnvelope) { e.GetRecordCollection().CustomerId = "invalid" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			var e pb.CommandEnvelope
			_ = proto.Unmarshal(body, &e)
			change(&e)
			raw, _ := proto.Marshal(&e)
			if _, _, err := codec.Decode(raw); err == nil {
				t.Fatal("invalid command accepted")
			}
		})
	}
}
