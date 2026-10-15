package messaging

import (
	"encoding/json"
	"fmt"
	internal "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/internalcommands"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging/generated"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/names"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"google.golang.org/protobuf/proto"
)

// RecordCollectionCodec owns the durable command format for the collection reaction.
type RecordCollectionCodec struct{}

func (RecordCollectionCodec) Owner() string { return "ordering" }
func (RecordCollectionCodec) Consumer() string {
	return names.OrderingSubscriptionRecordCollection.String()
}
func (c RecordCollectionCodec) Encode(m a.Metadata, command orderingcontracts.RecordCollectionCommand) ([]byte, error) {
	if err := internal.Validate(c.Owner(), c.Consumer(), m); err != nil {
		return nil, err
	}
	if err := d.ValidateID(command.CustomerID); err != nil {
		return nil, err
	}
	receipt, err := json.Marshal(m.Input)
	if err != nil {
		return nil, err
	}
	payload, err := ToRecordCollectionCommand(command)
	if err != nil {
		return nil, err
	}
	envelope := &pb.CommandEnvelope{Id: m.ID, Consumer: m.Consumer, SourceEventId: m.SourceEventID, SourceHash: m.SourceHash, Target: m.AggregateID, CorrelationId: m.CorrelationID, ReceiptMaterial: receipt, Command: &pb.CommandEnvelope_RecordCollection{RecordCollection: payload}}
	return proto.MarshalOptions{Deterministic: true}.Marshal(envelope)
}
func (c RecordCollectionCodec) Decode(body []byte) (a.Metadata, orderingcontracts.RecordCollectionCommand, error) {
	var envelope pb.CommandEnvelope
	if err := proto.Unmarshal(body, &envelope); err != nil {
		return a.Metadata{}, orderingcontracts.RecordCollectionCommand{}, err
	}
	payload := envelope.GetRecordCollection()
	if payload == nil {
		return a.Metadata{}, orderingcontracts.RecordCollectionCommand{}, fmt.Errorf("expected recordCollection private command")
	}
	m := a.Metadata{ID: envelope.Id, Consumer: envelope.Consumer, Name: envelope.Consumer, SourceEventID: envelope.SourceEventId, SourceHash: envelope.SourceHash, AggregateID: envelope.Target, CorrelationID: envelope.CorrelationId, CausationID: envelope.SourceEventId, Input: json.RawMessage(append([]byte(nil), envelope.ReceiptMaterial...))}
	command := FromRecordCollectionCommand(payload)
	if err := internal.Validate(c.Owner(), c.Consumer(), m); err != nil {
		return m, command, err
	}
	return m, command, d.ValidateID(command.CustomerID)
}
