// Package commandregistration connects command contracts directly to plain handler factories.
package commandregistration

import (
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/messaging/requests"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"google.golang.org/protobuf/proto"
)

// Registry supplies one aggregate's transaction to every registered command.
type Registry[A any] struct {
	requests    *rpc.RabbitMQRequestRegistry
	transaction execution.Transaction[A]
}

// New selects the owning request registry and aggregate transaction once.
func New[A any](requests *rpc.RabbitMQRequestRegistry, transaction execution.Transaction[A]) Registry[A] {
	return Registry[A]{requests, transaction}
}

// Register connects a wire command, parser and handler, adding shared execution policy.
func Register[A any, P proto.Message, C any, H execution.FeatureHandler[C]](
	registry Registry[A], name string, parse func(P) C, handler func(a.WriteRepository[A]) H,
) {
	rpc.Command(registry.requests, name, parse, execution.Bind(registry.transaction, handler).Execute)
}

// RegisterProjection prepares immutable projection data before opening the aggregate transaction.
func RegisterProjection[A any, P proto.Message, C, S any, H execution.FeatureHandler[C]](
	registry Registry[A], name string, parse func(P) C, source a.ProjectionPort[S], key func(C) string,
	handler func(a.WriteRepository[A], a.ProjectionPort[S]) H,
) {
	rpc.Command(registry.requests, name, parse, execution.BindProjection(registry.transaction, source, key, handler).Execute)
}
