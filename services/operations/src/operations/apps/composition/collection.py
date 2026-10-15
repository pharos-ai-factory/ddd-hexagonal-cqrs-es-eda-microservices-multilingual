"""Collection's providers bind its own database and durable command port."""

from operations.adaptors.generated.message_names import CollectionIntegrationEvent, CollectionSubscription

from operations.contexts.collection.adaptors.persistence.pickup_write_repository import (
    PostgresPickupWriteRepository,
)
from operations.contexts.collection.adaptors.persistence.pickup_snapshot import restore_pickup_snapshot
from operations.contexts.collection.adaptors.persistence.pickup_read_repository import (
    PostgresPickupReadRepository,
)
from operations.contexts.collection.adaptors.messaging.open_pickup_codec import command_codec
from operations.adaptors.generated.request_names import CollectionCommand, CollectionQuery
from operations.contexts.collection.application.queries.get_pickup import GetPickupQueryHandler
from operations.contexts.collection.application.queries.list_pickups import ListPickupsQueryHandler
from operations.contexts.collection.adaptors.messaging.pickup_publications import pickup_publications
from dependency_injector import containers, providers
from operations.apps.command_registration import CommandRegistration
from operations.adaptors.messaging.rabbitmq_request_registry import RabbitMQRequestRegistry
from operations.contexts.collection.adaptors.messaging import pickup_requests as wire
from operations.contexts.collection.adaptors.messaging.drinks_ready_subscription import DRINKS_READY
from operations.adaptors.persistence.aggregate_transaction import PostgresAggregateTransaction
from operations.adaptors.persistence.command_outbox import PostgresDurableCommandOutbox
from operations.contexts.collection.application.commands.open_pickup import (
    OpenPickupCommand,
    OpenPickupCommandHandler,
)
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommandHandler
from operations.contexts.collection.application.event_handlers.drinks_ready import (
    DrinksReadyIntegrationEventHandler,
)
from operations.contexts.collection.domain.pickup import PickupSnapshot, Pickup
from operations.foundation.identity import derived_id
from operations.apps.runtime import SubscriptionWorker, complete_subscriptions
from operations.apps.composition import context_database, subscription_definitions

from operations.adaptors.persistence.integration_event_outbox import PostgresPublicationOutbox
from operations.apps.runtime import publication_subscription
from operations.adaptors.generated.event_collection import publish_PickupOpened
from operations.contexts.collection.application.event_handlers.publish_pickup_opened import (
    PickupOpenedDomainEventHandler,
)
from operations.contexts.collection.adaptors.messaging.private_events import PICKUP_OPENED_DOMAIN
from operations.adaptors.generated.event_collection import publish_OrderCollected
from operations.contexts.collection.application.event_handlers.publish_order_collected import (
    OrderCollectedDomainEventHandler,
)
from operations.contexts.collection.adaptors.messaging.private_events import ORDER_COLLECTED_DOMAIN


class CollectionContainer(containers.DeclarativeContainer):
    """Owns the shared pool; application classes depend on transport-neutral ports."""

    config = providers.Configuration()
    database = providers.Resource(context_database, "collection", config.database_url)
    pickup_transaction = providers.Singleton(
        PostgresAggregateTransaction[PickupSnapshot, Pickup],
        database,
        "pickup",
        restore_pickup_snapshot,
        PostgresPickupWriteRepository,
        pickup_publications,
    )
    pickup_read_repository = providers.Singleton(PostgresPickupReadRepository, database)
    get_pickup = providers.Singleton(GetPickupQueryHandler, pickup_read_repository)
    list_pickups = providers.Singleton(ListPickupsQueryHandler, pickup_read_repository)
    codec = providers.Singleton(command_codec)
    outbox = providers.Singleton(PostgresDurableCommandOutbox[OpenPickupCommand], database, codec)
    ready = providers.Singleton(DrinksReadyIntegrationEventHandler, outbox)
    commands = providers.Singleton(CommandRegistration[Pickup], pickup_transaction)

    pickup_opened_publications = providers.Singleton(
        PostgresPublicationOutbox,
        database,
        CollectionSubscription.PUBLISH_PICKUP_OPENED,
        "pickup",
        publish_PickupOpened,
    )
    publish_pickup_opened = providers.Singleton(PickupOpenedDomainEventHandler, pickup_opened_publications)
    order_collected_publications = providers.Singleton(
        PostgresPublicationOutbox,
        database,
        CollectionSubscription.PUBLISH_ORDER_COLLECTED,
        "pickup",
        publish_OrderCollected,
    )
    publish_order_collected = providers.Singleton(
        OrderCollectedDomainEventHandler, order_collected_publications
    )


def requests(container: CollectionContainer) -> RabbitMQRequestRegistry:
    """Map published collection commands directly to their handlers."""
    registry = RabbitMQRequestRegistry("collection")
    container.commands().request(
        registry, CollectionCommand.COLLECT_ORDER, wire.collect, CollectOrderCommandHandler
    )
    registry.query(
        CollectionQuery.GET_PICKUP, wire.get_pickup, container.get_pickup().execute, wire.get_reply
    )
    registry.query(
        CollectionQuery.LIST_PICKUPS, wire.list_pickups, container.list_pickups().execute, wire.list_reply
    )
    return registry


def subscriptions(container: CollectionContainer) -> tuple[SubscriptionWorker, ...]:
    definitions = subscription_definitions("collection")
    return complete_subscriptions(
        "collection",
        definitions,
        container.commands().subscription(
            container.codec(),
            definitions,
            DRINKS_READY,
            lambda event: derived_id("pickup", event["orderId"]),
            container.ready().handle,
            lambda repository, realtime: OpenPickupCommandHandler(repository, realtime, derived_id),
        )
        + publication_subscription(
            "collection",
            CollectionSubscription.PUBLISH_PICKUP_OPENED,
            definitions,
            PICKUP_OPENED_DOMAIN,
            CollectionIntegrationEvent.PICKUP_OPENED,
            container.publish_pickup_opened().handle,
        )
        + publication_subscription(
            "collection",
            CollectionSubscription.PUBLISH_ORDER_COLLECTED,
            definitions,
            ORDER_COLLECTED_DOMAIN,
            CollectionIntegrationEvent.ORDER_COLLECTED,
            container.publish_order_collected().handle,
        ),
    )
