package support

import (
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/amqp"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"testing"
)

func TestSubscriptionGraphIsCompleteAndMatchesTheConsumerManifest(t *testing.T) {
	runtime := &StorefrontRuntime{Databases: map[string]*pg.ContextDatabase{"ordering": {}, "menu": {}}}
	if runtime.validateSubscriptions() == nil {
		t.Fatal("missing subscription accepted")
	}
	runtime.commandDecoders = map[string]map[string]a.Decoder{"ordering": {"ordering.record-collection": func([]byte) (a.Message, error) { return a.Message{}, nil }}}
	for _, binding := range expectedSubscriptions {
		sub := broker.Subscription{Binding: binding}
		if binding.Visibility == "command" {
			sub.DecodeCommand = func([]byte) (a.Metadata, a.Message, error) { return a.Metadata{}, a.Message{}, nil }
		}
		runtime.subscriptions = append(runtime.subscriptions, sub)
	}
	if err := runtime.validateSubscriptions(); err != nil {
		t.Fatal(err)
	}
	decoder := runtime.commandDecoders["ordering"]["ordering.record-collection"]
	delete(runtime.commandDecoders["ordering"], "ordering.record-collection")
	if runtime.validateSubscriptions() == nil {
		t.Fatal("missing command relay accepted")
	}
	runtime.commandDecoders["ordering"]["ordering.record-collection"] = decoder
	for i, sub := range runtime.subscriptions {
		if sub.Binding.Visibility != "command" {
			continue
		}
		original := runtime.subscriptions
		runtime.subscriptions = append(append([]broker.Subscription{}, original[:i]...), original[i+1:]...)
		if runtime.validateSubscriptions() == nil {
			t.Fatal("missing command consumer accepted")
		}
		runtime.subscriptions = original
		runtime.subscriptions[i].DecodeCommand = nil
		if runtime.validateSubscriptions() == nil {
			t.Fatal("missing private decoder accepted")
		}
		runtime.subscriptions[i] = sub
	}
	runtime.subscriptions = append(runtime.subscriptions, runtime.subscriptions[0])
	if runtime.validateSubscriptions() == nil {
		t.Fatal("duplicate subscription accepted")
	}
	runtime.subscriptions = runtime.subscriptions[:len(runtime.subscriptions)-1]
	runtime.subscriptions[0].Binding.Event = "invalid.event"
	if runtime.validateSubscriptions() == nil {
		t.Fatal("wrong event accepted")
	}
}
