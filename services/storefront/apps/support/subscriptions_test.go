package support

import (
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/persistence/postgres"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/transport/amqp"
	"testing"
)

func TestSubscriptionGraphIsCompleteAndMatchesTheConsumerManifest(t *testing.T) {
	runtime := &StorefrontRuntime{Databases: map[string]*pg.ContextDatabase{"ordering": {}}}
	if runtime.validateSubscriptions() == nil {
		t.Fatal("missing subscription accepted")
	}
	binding := expectedSubscriptions["ordering.menu-directory"]
	runtime.subscriptions = []broker.Subscription{{Binding: binding}}
	if err := runtime.validateSubscriptions(); err != nil {
		t.Fatal(err)
	}
	runtime.subscriptions = append(runtime.subscriptions, runtime.subscriptions[0])
	if runtime.validateSubscriptions() == nil {
		t.Fatal("duplicate subscription accepted")
	}
	runtime.subscriptions = runtime.subscriptions[:1]
	runtime.subscriptions[0].Binding.Event = "ordering.order-placed"
	if runtime.validateSubscriptions() == nil {
		t.Fatal("wrong event accepted")
	}
}
