package support

import (
	"context"
	"fmt"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/amqp"
	internal "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/internalcommands"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// SubscribeCommandPair installs acceptance, command execution and its registered relay decoder together.
func SubscribeCommandPair[P, C any](s *StorefrontRuntime, binding broker.Binding, codec internal.Codec[C], target func(P) string, react func(context.Context, a.Metadata, P) (a.Outcome, error), execute func(context.Context, a.Metadata, C) (a.Outcome, error)) {
	if !binding.Command || binding.Consumer != codec.Consumer() || binding.Context != codec.Owner() {
		panic("command subscription disagrees with codec")
	}
	if s.commandDecoders == nil {
		s.commandDecoders = map[string]map[string]a.Decoder{}
	}
	if s.commandDecoders[binding.Context] == nil {
		s.commandDecoders[binding.Context] = map[string]a.Decoder{}
	}
	if s.commandDecoders[binding.Context][binding.Consumer] != nil {
		panic("duplicate command subscription")
	}
	decode := internal.DeliveryDecoder(codec)
	s.commandDecoders[binding.Context][binding.Consumer] = func(body []byte) (a.Message, error) { _, message, err := decode(body); return message, err }
	if binding.Visibility == string(a.Private) {
		SubscribePrivate(s, binding, target, react)
	} else {
		Subscribe(s, binding, target, react)
	}
	consumer := binding.Consumer + ".command"
	s.subscriptions = append(s.subscriptions, broker.Subscription{
		Binding: broker.Binding{Consumer: consumer, Event: consumer, Visibility: "command", Context: binding.Context}, DecodeCommand: decode,
		Handle: func(ctx context.Context, m a.Metadata, message a.Message) error {
			command, ok := message.Payload.(C)
			if !ok {
				return fmt.Errorf("private command payload type mismatch")
			}
			outcome, err := execute(ctx, m, command)
			if err != nil {
				return err
			}
			if outcome.Rejection != nil {
				return outcome.Rejection
			}
			return nil
		},
	})
}
