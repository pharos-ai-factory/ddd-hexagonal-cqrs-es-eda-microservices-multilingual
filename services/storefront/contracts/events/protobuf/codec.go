package protobuf

import (
	"fmt"
	private "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/generated"
	menu "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/cafe/v1"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"google.golang.org/protobuf/proto"
	"time"
)

func Encode(message a.Message) ([]byte, error) {
	if err := validateEnvelope(message); err != nil {
		return nil, err
	}
	event := &pb.Event{Id: message.ID, Name: message.Name, Context: message.Context, Visibility: string(message.Visibility), ContractVersion: message.ContractVersion, AggregateKind: message.AggregateKind, AggregateId: message.AggregateID, AggregateVersion: message.AggregateVersion, CorrelationId: message.CorrelationID, CausationId: message.CausationID, OccurredAt: message.OccurredAt.UTC().Format(time.RFC3339Nano)}
	if message.Visibility == a.Private {
		return encodePrivate(message)
	}
	expected := ""
	switch p := message.Payload.(type) {
	case model.MenuPublished:
		expected = "menu.edition-published"
		event.Payload = &pb.Event_MenuPublished{MenuPublished: &pb.MenuPublished{EditionId: p.EditionID, Currency: p.Currency, Offers: encodeOffers(p.Offers)}}
	case model.OrderPlaced:
		expected = "ordering.order-placed"
		event.Payload = &pb.Event_OrderPlaced{OrderPlaced: &pb.OrderPlaced{OrderId: p.OrderID, CustomerId: p.CustomerID, EditionId: p.EditionID, Currency: p.Currency, Lines: encodeLines(p.Lines)}}
	case model.DrinksReady:
		expected = "preparation.drinks-ready"
		event.Payload = &pb.Event_DrinksReady{DrinksReady: &pb.DrinksReady{OrderId: p.OrderID, CustomerId: p.CustomerID}}
	case model.PickupOpened:
		expected = "collection.pickup-opened"
		event.Payload = &pb.Event_PickupOpened{PickupOpened: &pb.PickupOpened{PickupId: p.PickupID, OrderId: p.OrderID, CustomerId: p.CustomerID, CollectionCode: p.CollectionCode}}
	case model.OrderCollected:
		expected = "collection.order-collected"
		event.Payload = &pb.Event_OrderCollected{OrderCollected: &pb.OrderCollected{OrderId: p.OrderID, CustomerId: p.CustomerID}}
	case model.RewardIssued:
		expected = "loyalty.reward-issued"
		event.Payload = &pb.Event_RewardIssued{RewardIssued: &pb.RewardIssued{RewardId: p.RewardID, CustomerId: p.CustomerID, Benefit: p.Benefit, ExpiresAt: p.ExpiresAt}}
	default:
		return nil, fmt.Errorf("unknown publication payload %T", message.Payload)
	}
	if expected != message.Name {
		return nil, fmt.Errorf("event name and typed payload disagree")
	}
	return proto.MarshalOptions{Deterministic: true}.Marshal(event)
}
func Decode(data []byte) (a.Message, error) {
	var event pb.Event
	if len(data) > 256*1024 {
		return a.Message{}, fmt.Errorf("event exceeds size limit")
	}
	if err := proto.Unmarshal(data, &event); err != nil {
		return a.Message{}, err
	}
	if event.Visibility == string(a.Private) {
		return decodePrivate(data)
	}
	if len(event.ProtoReflect().GetUnknown()) > 0 {
		return a.Message{}, fmt.Errorf("unknown envelope fields")
	}
	occurred, err := time.Parse(time.RFC3339Nano, event.OccurredAt)
	if err != nil {
		return a.Message{}, err
	}
	message := a.Message{ID: event.Id, Name: event.Name, Context: event.Context, Visibility: a.Visibility(event.Visibility), ContractVersion: event.ContractVersion, AggregateKind: event.AggregateKind, AggregateID: event.AggregateId, AggregateVersion: event.AggregateVersion, CorrelationID: event.CorrelationId, CausationID: event.CausationId, OccurredAt: occurred}
	expected := ""
	switch value := event.Payload.(type) {
	case *pb.Event_MenuPublished:
		p := value.MenuPublished
		if p == nil {
			return a.Message{}, fmt.Errorf("missing payload")
		}
		expected = "menu.edition-published"
		message.Payload = model.MenuPublished{EditionID: p.EditionId, Currency: p.Currency, Offers: decodeOffers(p.Offers)}
	case *pb.Event_OrderPlaced:
		p := value.OrderPlaced
		if p == nil {
			return a.Message{}, fmt.Errorf("missing payload")
		}
		expected = "ordering.order-placed"
		message.Payload = model.OrderPlaced{OrderID: p.OrderId, CustomerID: p.CustomerId, EditionID: p.EditionId, Currency: p.Currency, Lines: decodeLines(p.Lines)}
	case *pb.Event_DrinksReady:
		p := value.DrinksReady
		if p == nil {
			return a.Message{}, fmt.Errorf("missing payload")
		}
		expected = "preparation.drinks-ready"
		message.Payload = model.DrinksReady{OrderID: p.OrderId, CustomerID: p.CustomerId}
	case *pb.Event_PickupOpened:
		p := value.PickupOpened
		if p == nil {
			return a.Message{}, fmt.Errorf("missing payload")
		}
		expected = "collection.pickup-opened"
		message.Payload = model.PickupOpened{PickupID: p.PickupId, OrderID: p.OrderId, CustomerID: p.CustomerId, CollectionCode: p.CollectionCode}
	case *pb.Event_OrderCollected:
		p := value.OrderCollected
		if p == nil {
			return a.Message{}, fmt.Errorf("missing payload")
		}
		expected = "collection.order-collected"
		message.Payload = model.OrderCollected{OrderID: p.OrderId, CustomerID: p.CustomerId}
	case *pb.Event_RewardIssued:
		p := value.RewardIssued
		if p == nil {
			return a.Message{}, fmt.Errorf("missing payload")
		}
		expected = "loyalty.reward-issued"
		message.Payload = model.RewardIssued{RewardID: p.RewardId, CustomerID: p.CustomerId, Benefit: p.Benefit, ExpiresAt: p.ExpiresAt}
	default:
		return a.Message{}, fmt.Errorf("missing or unknown payload")
	}
	if expected != message.Name {
		return a.Message{}, fmt.Errorf("event name and payload disagree")
	}
	if err := validateEnvelope(message); err != nil {
		return a.Message{}, err
	}
	return message, nil
}
func validateEnvelope(m a.Message) error {
	definition, ok := model.Lookup(m.Name)
	if m.Visibility == a.Private {
		ok = m.Name == "menu.drink-published" && m.Context == "menu" && m.AggregateKind == "drink"
	} else {
		ok = ok && definition.Owner == m.Context && definition.Visibility == m.Visibility && m.AggregateKind == definition.AggregateKind
	}
	if !ok || m.ContractVersion != 1 {
		return fmt.Errorf("unknown event contract or invalid owner/visibility")
	}
	for _, id := range []string{m.ID, m.AggregateID, m.CorrelationID, m.CausationID} {
		if err := core.ValidateID(id); err != nil {
			return err
		}
	}
	if m.AggregateVersion == 0 || m.AggregateKind == "" || m.OccurredAt.IsZero() {
		return fmt.Errorf("incomplete aggregate metadata")
	}
	return validatePayload(m)
}
func encodeOffers(values []model.Offer) []*pb.Offer {
	out := make([]*pb.Offer, 0, len(values))
	for _, v := range values {
		out = append(out, &pb.Offer{Code: v.Code, DrinkId: v.DrinkID, DrinkRevision: v.DrinkRevision, Name: v.Name, Minor: v.Minor, Currency: v.Currency})
	}
	return out
}
func decodeOffers(values []*pb.Offer) []model.Offer {
	out := make([]model.Offer, 0, len(values))
	for _, v := range values {
		if v != nil {
			out = append(out, model.Offer{Code: v.Code, DrinkID: v.DrinkId, DrinkRevision: v.DrinkRevision, Name: v.Name, Minor: v.Minor, Currency: v.Currency})
		}
	}
	return out
}
func encodeLines(values []model.Line) []*pb.Line {
	out := make([]*pb.Line, 0, len(values))
	for _, v := range values {
		out = append(out, &pb.Line{Id: v.ID, OfferCode: v.OfferCode, Name: v.Name, Quantity: int32(v.Quantity), Minor: v.Minor})
	}
	return out
}
func decodeLines(values []*pb.Line) []model.Line {
	out := make([]model.Line, 0, len(values))
	for _, v := range values {
		if v != nil {
			out = append(out, model.Line{ID: v.Id, OfferCode: v.OfferCode, Name: v.Name, Quantity: int(v.Quantity), Minor: v.Minor})
		}
	}
	return out
}

func encodePrivate(message a.Message) ([]byte, error) {
	p, ok := message.Payload.(menu.DrinkPublished)
	if !ok || message.Name != "menu.drink-published" || message.Context != "menu" {
		return nil, fmt.Errorf("foreign private message")
	}
	event := &private.PrivateEvent{Id: message.ID, Name: message.Name, Context: message.Context, Visibility: "domain", FormatRevision: message.ContractVersion, AggregateKind: message.AggregateKind, AggregateId: message.AggregateID, AggregateVersion: message.AggregateVersion, CorrelationId: message.CorrelationID, CausationId: message.CausationID, OccurredAt: message.OccurredAt.UTC().Format(time.RFC3339Nano), Payload: &private.PrivateEvent_DrinkPublished{DrinkPublished: &private.DrinkPublished{DrinkId: p.DrinkID, Name: p.Name, Revision: p.Revision}}}
	return proto.MarshalOptions{Deterministic: true}.Marshal(event)
}
func decodePrivate(data []byte) (a.Message, error) {
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
	message := a.Message{ID: event.Id, Name: event.Name, Context: event.Context, Visibility: a.Visibility(event.Visibility), ContractVersion: event.FormatRevision, AggregateKind: event.AggregateKind, AggregateID: event.AggregateId, AggregateVersion: event.AggregateVersion, CorrelationID: event.CorrelationId, CausationID: event.CausationId, OccurredAt: occurred, Payload: menu.DrinkPublished{DrinkID: p.DrinkId, Name: p.Name, Revision: p.Revision}}
	if message.Visibility != a.Private {
		return a.Message{}, fmt.Errorf("private message escaped its owner")
	}
	return message, validateEnvelope(message)
}
