package commandregistration

import (
	"context"
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	broker "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/amqp"
	internal "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/internalcommands"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Subscribe connects the owner codec and plain handler factory to both durable consumers.
func Subscribe[A, P, C any, H execution.FeatureHandler[C]](registry Registry[A], s *support.StorefrontRuntime, binding broker.Binding, codec internal.Codec[C], target func(P) string, react func(context.Context, a.Metadata, P) (a.Outcome, error), handler func(a.WriteRepository[A], a.RealtimeBusPort[a.RealtimeEvent]) H) {
	support.SubscribeCommandPair(s, binding, codec, target, react, execution.Bind(registry.transaction, handler).Execute)
}
