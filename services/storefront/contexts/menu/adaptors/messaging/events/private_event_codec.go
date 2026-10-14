// Package events owns private event formats and validates their domain evidence.
package events

import (
	"fmt"
	private "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/generated"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/envelope"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/names"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"time"
)

// EncodePrivateEvent preserves owner-private wire identities and typed facts.
func EncodePrivateEvent(message a.Message) ([]byte, error) {
	if message.Name == names.MenuDomainEventDrinkPublished.String() {
		return encodeDrink(message)
	}
	value, ok := message.Payload.(d.MenuEditionPublishedDomainEvent)
	if !ok {
		return nil, fmt.Errorf("wrong private payload")
	}
	if err := validateMenuEditionPublished(message, value); err != nil {
		return nil, err
	}
	event := &private.PrivateEvent{Id: message.ID, Name: message.Name, Context: message.Context, Visibility: "domain", FormatRevision: message.ContractVersion, AggregateKind: message.AggregateKind, AggregateId: message.AggregateID, AggregateVersion: message.AggregateVersion, CorrelationId: message.CorrelationID, CausationId: message.CausationID, OccurredAt: message.OccurredAt.UTC().Format(time.RFC3339Nano)}
	s := value.State
	payload := &private.MenuEditionPublishedDomainEvent{EditionId: s.ID, Currency: s.Currency}
	for _, offer := range s.Offers {
		payload.Offers = append(payload.Offers, &private.Offer{Code: offer.Code, DrinkId: offer.DrinkID, DrinkRevision: offer.DrinkRevision, Name: offer.Name, Minor: offer.Minor, Currency: offer.Currency})
	}
	event.Payload = &private.PrivateEvent_EditionPublished{EditionPublished: payload}
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
	if event.Name == names.MenuDomainEventDrinkPublished.String() {
		return decodeDrink(data)
	}
	p := event.GetEditionPublished()
	if p == nil || len(event.ProtoReflect().GetUnknown()) > 0 {
		return a.Message{}, fmt.Errorf("unknown private payload")
	}
	occurred, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil {
		return a.Message{}, err
	}
	s := d.EditionState{ID: p.EditionId, Currency: p.Currency, Status: "published", Offers: []d.OfferState{}}
	for _, offer := range p.Offers {
		s.Offers = append(s.Offers, d.OfferState{Code: offer.Code, DrinkID: offer.DrinkId, DrinkRevision: offer.DrinkRevision, Name: offer.Name, Minor: offer.Minor, Currency: offer.Currency})
	}
	value := d.MenuEditionPublishedDomainEvent{State: s}
	message := a.Message{ID: event.Id, Name: event.Name, Context: event.Context, Visibility: a.Visibility(event.Visibility), ContractVersion: event.FormatRevision, AggregateKind: event.AggregateKind, AggregateID: event.AggregateId, AggregateVersion: event.AggregateVersion, CorrelationID: event.CorrelationId, CausationID: event.CausationId, OccurredAt: occurred, Payload: value}
	return message, validateMenuEditionPublished(message, value)
}
func validateMenuEditionPublished(m a.Message, value d.MenuEditionPublishedDomainEvent) error {
	if err := envelope.ValidateMetadata(m); err != nil {
		return err
	}
	if m.Context != "menu" || m.Name != names.MenuDomainEventEditionPublished.String() || m.Visibility != a.Private || m.AggregateKind != "edition" || m.ContractVersion != 1 || value.State.ID != m.AggregateID || value.State.Status != "published" {
		return fmt.Errorf("invalid private menu event")
	}
	_, err := d.RestoreEdition(value.State)
	return err
}
