package protobuf

import (
	"fmt"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/cafe/v1"
	core "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"regexp"
	"strings"
	"time"
)

// validatePayload applies the same declared structural constraints to every payload.
func validatePayload(value protoreflect.Message, source string) error {
	if !value.IsValid() {
		return fmt.Errorf("missing payload message")
	}
	fields := value.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		rule, _ := proto.GetExtension(field.Options(), pb.E_Rules).(*pb.FieldRules)
		if rule == nil {
			rule = &pb.FieldRules{}
		}
		item := value.Get(field)
		if field.IsList() {
			list := item.List()
			if rule.MinItems != nil && uint32(list.Len()) < *rule.MinItems || rule.MaxItems != nil && uint32(list.Len()) > *rule.MaxItems {
				return fmt.Errorf("invalid collection size: %s", field.Name())
			}
			for j := 0; j < list.Len(); j++ {
				if err := validatePayload(list.Get(j).Message(), source); err != nil {
					return err
				}
			}
			continue
		}
		switch field.Kind() {
		case protoreflect.MessageKind:
			if !value.Has(field) {
				return fmt.Errorf("missing message: %s", field.Name())
			}
			if err := validatePayload(item.Message(), source); err != nil {
				return err
			}
		case protoreflect.StringKind:
			text := item.String()
			if rule.Uuid {
				if err := core.ValidateID(text); err != nil {
					return err
				}
			}
			if rule.SourceIdentity && text != source {
				return fmt.Errorf("payload does not identify its source aggregate")
			}
			if rule.NonEmpty && strings.TrimSpace(text) == "" {
				return fmt.Errorf("empty string: %s", field.Name())
			}
			if rule.Pattern != "" {
				matched, err := regexp.MatchString(rule.Pattern, text)
				if err != nil || !matched {
					return fmt.Errorf("invalid string: %s", field.Name())
				}
			}
			if rule.Timestamp {
				if _, err := time.Parse(time.RFC3339Nano, text); err != nil {
					return err
				}
			}
		case protoreflect.BoolKind:
		default:
			var number int64
			if field.Kind() == protoreflect.Uint64Kind || field.Kind() == protoreflect.Uint32Kind {
				if item.Uint() > 9007199254740991 {
					return fmt.Errorf("integer outside portable range")
				}
				number = int64(item.Uint())
			} else {
				number = item.Int()
			}
			if number < -9007199254740991 || number > 9007199254740991 || rule.Minimum != nil && number < *rule.Minimum || rule.Maximum != nil && number > *rule.Maximum {
				return fmt.Errorf("invalid integer: %s", field.Name())
			}
		}
	}
	return nil
}
