// Package events owns private event formats and validates their domain evidence.
package events

import (
	"fmt"
	private "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging/generated"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/envelope"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/names"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"time"
)

// EncodePrivateEvent preserves owner-private wire identities and typed facts.
func EncodePrivateEvent(message a.Message) ([]byte, error) {

	value, ok := message.Payload.(d.OrderPlacedDomainEvent)
	if !ok {
		return nil, fmt.Errorf("wrong private payload")
	}
	if err := validateOrderPlaced(message, value); err != nil {
		return nil, err
	}
	event := &private.PrivateEvent{Id: message.ID, Name: message.Name, Context: message.Context, Visibility: "domain", FormatRevision: message.ContractVersion, AggregateKind: message.AggregateKind, AggregateId: message.AggregateID, AggregateVersion: message.AggregateVersion, CorrelationId: message.CorrelationID, CausationId: message.CausationID, OccurredAt: message.OccurredAt.UTC().Format(time.RFC3339Nano)}
	s := value.State
	payload := &private.OrderPlacedDomainEvent{OrderId: s.ID, CustomerId: s.CustomerID, EditionId: s.EditionID, Currency: s.Currency}
	for _, line := range s.Lines {
		payload.Lines = append(payload.Lines, &private.Line{Id: line.ID, OfferCode: line.Selection.OfferCode, Name: line.Selection.Name, Quantity: int32(line.Quantity), Minor: line.Selection.Minor})
	}
	event.Payload = &private.PrivateEvent_OrderPlaced{OrderPlaced: payload}
	body, err := proto.MarshalOptions{Deterministic: true}.Marshal(event)
	if len(body) > 256*1024 {
		return nil, fmt.Errorf("event exceeds wire limit")
	}
	return body, err
}

// DecodePrivateEvent restores plain facts after checking scope and immutable evidence.
func DecodePrivateEvent(data []byte) (a.Message, error) {
	event := &private.PrivateEvent{}
	if err := proto.Unmarshal(data, event); err != nil {
		return a.Message{}, err
	}

	p := event.GetOrderPlaced()
	if p == nil || len(event.ProtoReflect().GetUnknown()) > 0 {
		return a.Message{}, fmt.Errorf("unknown private payload")
	}
	occurred, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil {
		return a.Message{}, err
	}
	s := d.OrderState{ID: p.OrderId, CustomerID: p.CustomerId, EditionID: p.EditionId, Currency: p.Currency, Status: "placed", Lines: []d.LineState{}}
	for _, line := range p.Lines {
		s.Lines = append(s.Lines, d.LineState{ID: line.Id, Selection: d.Selection{OfferCode: line.OfferCode, Name: line.Name, Minor: line.Minor}, Quantity: int(line.Quantity)})
	}
	value := d.OrderPlacedDomainEvent{State: s}
	message := a.Message{ID: event.Id, Name: event.Name, Context: event.Context, Visibility: a.Visibility(event.Visibility), ContractVersion: event.FormatRevision, AggregateKind: event.AggregateKind, AggregateID: event.AggregateId, AggregateVersion: event.AggregateVersion, CorrelationID: event.CorrelationId, CausationID: event.CausationId, OccurredAt: occurred, Payload: value}
	return message, validateOrderPlaced(message, value)
}
func validateOrderPlaced(m a.Message, value d.OrderPlacedDomainEvent) error {
	if err := envelope.ValidateMetadata(m); err != nil {
		return err
	}
	if m.Context != "ordering" || m.Name != names.OrderingDomainEventOrderPlaced.String() || m.Visibility != a.Private || m.AggregateKind != "order" || m.ContractVersion != 1 || value.State.ID != m.AggregateID || value.State.Status != "placed" {
		return fmt.Errorf("invalid private ordering event")
	}
	_, err := d.Restore(value.State)
	return err
}
