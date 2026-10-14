package di

import (
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/messaging"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/adaptors/persistence/postgres"
	reactions "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/eventhandlers"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	orderingprojections "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/projections"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/queries"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/names"
	codec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"go.uber.org/fx"
)

func orderingModule() fx.Option {
	return fx.Module("ordering", fx.Provide(
		func(s *support.StorefrontRuntime) reactions.OrderPlacedDomainEventHandler {
			return reactions.OrderPlacedDomainEventHandler{Publications: store.NewPublicationOutbox[model.OrderPlacedIntegrationEvent](s.Databases["ordering"], names.OrderingSubscriptionPublishOrderPlaced.String(), "order", codec.PublishOrderPlaced)}
		},
		func(s *support.StorefrontRuntime) execution.Transaction[*d.Order] {
			return pg.NewOrderTransaction(s.Databases["ordering"])
		},
		func(s *support.StorefrontRuntime) ports.OrderReadRepository {
			return pg.NewOrderReadRepository(s.Databases["ordering"])
		},
		func(s *support.StorefrontRuntime) a.ProjectionPort[model.MenuPublishedIntegrationEvent] {
			return store.Project[model.MenuPublishedIntegrationEvent](s.Databases["ordering"], "published-menus")
		},
		func(read ports.OrderReadRepository) q.GetOrderQueryHandler {
			return q.GetOrderQueryHandler{ReadRepository: read}
		},
		func(read ports.OrderReadRepository) q.ListOrdersQueryHandler {
			return q.ListOrdersQueryHandler{ReadRepository: read}
		},
		orderingRequestsProvider,
	), fx.Invoke(mountOrdering))
}

func mountOrdering(s *support.StorefrontRuntime, requests orderingRequests, directory a.ProjectionPort[model.MenuPublishedIntegrationEvent], publication reactions.OrderPlacedDomainEventHandler) {
	support.SubscribePrivate(s, rpc.Subscription(names.OrderingSubscriptionPublishOrderPlaced), func(event d.OrderPlacedDomainEvent) string { return event.State.ID }, publication.Handle)
	s.RequestWorkers["ordering"] = requests.Run
	support.Subscribe(s, rpc.Subscription(names.OrderingSubscriptionMenuDirectory), func(p model.MenuPublishedIntegrationEvent) string { return p.EditionID }, rpc.MenuPublished(orderingprojections.MenuDirectoryProjectionHandler{Directory: directory}.Handle))
}
