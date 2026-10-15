// Package internalcommands defines the technical boundary for owner-private command codecs.
package internalcommands

import (
	"encoding/json"
	"fmt"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"regexp"
	"strings"
)

// Codec maps exact private bytes to plain owner metadata and a command DTO.
type Codec[C any] interface {
	Owner() string
	Consumer() string
	Encode(a.Metadata, C) ([]byte, error)
	Decode([]byte) (a.Metadata, C, error)
}

var sourceHash = regexp.MustCompile(`^[a-f0-9]{64}$`)

// Validate preserves source identity and receipt material across durable acceptance.
func Validate(owner, consumer string, m a.Metadata) error {
	for _, id := range []string{m.ID, m.AggregateID, m.SourceEventID, m.CorrelationID} {
		if err := d.ValidateID(id); err != nil {
			return err
		}
	}
	if !strings.HasPrefix(consumer, owner+".") || m.Consumer != consumer || m.Name != consumer || m.ID != a.DerivedID(consumer, m.SourceEventID) || m.CausationID != m.SourceEventID || m.ExpectedVersion != nil || !sourceHash.MatchString(m.SourceHash) {
		return fmt.Errorf("invalid private command receipt metadata")
	}
	body, err := json.Marshal(m.Input)
	if err != nil {
		return err
	}
	var object map[string]json.RawMessage
	if err = json.Unmarshal(body, &object); err != nil || object == nil {
		return fmt.Errorf("private command receipt material must be an object")
	}
	return nil
}

// DeliveryDecoder preserves command metadata instead of deriving a second command identity.
func DeliveryDecoder[C any](codec Codec[C]) func([]byte) (a.Metadata, a.Message, error) {
	return func(body []byte) (a.Metadata, a.Message, error) {
		m, c, err := codec.Decode(body)
		if err == nil {
			err = Validate(codec.Owner(), codec.Consumer(), m)
		}
		return m, a.Message{ID: m.ID, Name: codec.Consumer() + ".command", Context: codec.Owner(), Visibility: "command", ContractVersion: 1, AggregateID: m.AggregateID, CorrelationID: m.CorrelationID, CausationID: m.SourceEventID, Payload: c}, err
	}
}
