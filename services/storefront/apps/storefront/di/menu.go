package di

import (
	execution "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/command"
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/apps/support"
	rpc "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging"
	endpoints "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/messaging/queryendpoints"
	pg "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/adaptors/persistence/postgres"
	app "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application"
	reactions "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/eventhandlers"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/ports"
	menuprojections "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/projections"
	q "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/application/queries"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/menu/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	codec "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/protobuf"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"go.uber.org/fx"
)

func menuModule() fx.Option {
	return fx.Module("menu", fx.Provide(
		func(s *support.StorefrontRuntime) reactions.MenuEditionPublishedDomainEventHandler {
			return reactions.MenuEditionPublishedDomainEventHandler{Publications: store.NewPublicationOutbox[model.MenuPublishedIntegrationEvent](s.Databases["menu"], "menu.publish-edition-published", "edition", codec.PublishMenuPublished)}
		},
		func(s *support.StorefrontRuntime) execution.Transaction[*d.Drink] {
			return pg.NewDrinkTransaction(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) execution.Transaction[*d.MenuEdition] {
			return pg.NewEditionTransaction(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) ports.DrinkReadRepository {
			return pg.NewDrinkReadRepository(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) ports.EditionReadRepository {
			return pg.NewEditionReadRepository(s.Databases["menu"])
		},
		func(s *support.StorefrontRuntime) a.ProjectionPort[app.DrinkPublishedDomainEvent] {
			return store.Project[app.DrinkPublishedDomainEvent](s.Databases["menu"], "published-drinks")
		},
		func(read ports.DrinkReadRepository) q.GetDrinkQueryHandler {
			return q.GetDrinkQueryHandler{ReadRepository: read}
		},
		func(read ports.DrinkReadRepository) q.ListDrinksQueryHandler {
			return q.ListDrinksQueryHandler{ReadRepository: read}
		},
		func(get q.GetDrinkQueryHandler, list q.ListDrinksQueryHandler) endpoints.DrinkQueryEndpoints {
			return endpoints.DrinkQueryEndpoints{GetHandler: get, ListHandler: list}
		},
		func(read ports.EditionReadRepository) q.GetEditionQueryHandler {
			return q.GetEditionQueryHandler{ReadRepository: read}
		},
		func(read ports.EditionReadRepository) q.ListEditionsQueryHandler {
			return q.ListEditionsQueryHandler{ReadRepository: read}
		},
		func(get q.GetEditionQueryHandler, list q.ListEditionsQueryHandler) endpoints.EditionQueryEndpoints {
			return endpoints.EditionQueryEndpoints{GetHandler: get, ListHandler: list}
		},
		menuRequestsProvider,
	), fx.Invoke(mountMenu))
}

func mountMenu(s *support.StorefrontRuntime, requests menuRequests, directory a.ProjectionPort[app.DrinkPublishedDomainEvent], publication reactions.MenuEditionPublishedDomainEventHandler) {
	support.SubscribePrivate(s, rpc.Subscription(string(PublishMenuEditionPublishedSubscription)), func(event d.MenuEditionPublishedDomainEvent) string { return event.State.ID }, publication.Handle)
	s.RequestWorkers["menu"] = requests.Run
	support.SubscribePrivate(s, rpc.Subscription(string(DrinkDirectorySubscription)), func(p app.DrinkPublishedDomainEvent) string { return p.DrinkID }, menuprojections.DrinkDirectoryProjectionHandler{Directory: directory}.Handle)
}

// MenuSubscription identifies private projection delivery without untyped routing literals.
type MenuSubscription string

const DrinkDirectorySubscription MenuSubscription = "menu.drink-directory"

// PublishMenuEditionPublishedSubscription independently delivers the private publication reaction.
const PublishMenuEditionPublishedSubscription MenuSubscription = "menu.publish-edition-published"
