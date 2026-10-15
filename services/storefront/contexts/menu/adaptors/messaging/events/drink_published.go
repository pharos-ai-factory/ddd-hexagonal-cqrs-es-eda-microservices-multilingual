package events

import (
	"fmt"
	private "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/generated"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	envelope "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/envelope"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/names"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
	"unicode/utf8"
)

func encodeDrink(message a.Message) ([]byte, error) {
	if err := validatePrivate(message); err != nil {
		return nil, err
	}
	p, ok := message.Payload.(menu.DrinkPublishedDomainEvent)
	if !ok || message.Name != names.MenuDomainEventDrinkPublished.String() || message.Context != "menu" {
		return nil, fmt.Errorf("foreign private message")
	}
	event := &private.PrivateEvent{Id: message.ID, Name: message.Name, Context: message.Context, Visibility: "domain", FormatRevision: message.ContractVersion, AggregateKind: message.AggregateKind, AggregateId: message.AggregateID, AggregateVersion: message.AggregateVersion, CorrelationId: message.CorrelationID, CausationId: message.CausationID, OccurredAt: message.OccurredAt.UTC().Format(time.RFC3339Nano), Payload: &private.PrivateEvent_DrinkPublished{DrinkPublished: &private.DrinkPublished{DrinkId: p.DrinkID, Name: p.Name, Revision: p.Revision}}}
	return proto.MarshalOptions{Deterministic: true}.Marshal(event)
}
func decodeDrink(data []byte) (a.Message, error) {
	event := &private.PrivateEvent{}
	if err := proto.Unmarshal(data, event); err != nil {
		return a.Message{}, err
	}
	if len(event.ProtoReflect().GetUnknown()) > 0 || event.GetDrinkPublished() == nil {
		return a.Message{}, fmt.Errorf("unknown private message format")
	}
	occurred, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil {
		return a.Message{}, err
	}
	p := event.GetDrinkPublished()
	message := a.Message{ID: event.Id, Name: event.Name, Context: event.Context, Visibility: a.Visibility(event.Visibility), ContractVersion: event.FormatRevision, AggregateKind: event.AggregateKind, AggregateID: event.AggregateId, AggregateVersion: event.AggregateVersion, CorrelationID: event.CorrelationId, CausationID: event.CausationId, OccurredAt: occurred, Payload: menu.DrinkPublishedDomainEvent{DrinkID: p.DrinkId, Name: p.Name, Revision: p.Revision}}
	if message.Visibility != a.Private {
		return a.Message{}, fmt.Errorf("private message escaped its owner")
	}
	return message, validatePrivate(message)
}

func validatePrivate(m a.Message) error {
	if err := envelope.ValidateMetadata(m); err != nil {
		return err
	}
	p, ok := m.Payload.(menu.DrinkPublishedDomainEvent)
	if !ok || m.Name != names.MenuDomainEventDrinkPublished.String() || m.Context != "menu" || m.Visibility != a.Private || m.AggregateKind != "drink" || m.ContractVersion != 1 || p.DrinkID != m.AggregateID || strings.TrimSpace(p.Name) == "" || !utf8.ValidString(p.Name) || utf8.RuneCountInString(p.Name) > 80 || p.Revision == 0 {
		return fmt.Errorf("invalid private drink publication")
	}
	return nil
}
