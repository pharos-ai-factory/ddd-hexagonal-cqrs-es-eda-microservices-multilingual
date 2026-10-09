"""Collection's providers bind its own database and durable command port."""
from operations.contexts.collection.adaptors.persistence.pickup_write_repository import PostgresPickupWriteRepository
from operations.contexts.collection.adaptors.persistence.pickup_snapshot import restore_pickup_snapshot
from operations.contexts.collection.adaptors.persistence.pickup_read_repository import PostgresPickupReadRepository
from operations.contexts.collection.adaptors.messaging.open_pickup_codec import command_codec
from operations.contexts.collection.adaptors.query_endpoints import PickupQueryEndpoints
from operations.contexts.collection.application.queries.get_pickup import GetPickupQueryHandler
from operations.contexts.collection.application.queries.list_pickups import ListPickupsQueryHandler
from operations.contexts.collection.adaptors.messaging.pickup_publications import pickup_publications
from dependency_injector import containers, providers
from operations.adaptors.command_execution import CommandExecutor
from operations.contexts.collection.adaptors.messaging.incoming_event import DRINKS_READY
from operations.adaptors.aggregate_transaction import PostgresAggregateTransaction
from operations.adaptors.internal_commands import PostgresDurableCommandOutbox
from operations.contexts.collection.application.commands.open_pickup import OpenPickupCommand, OpenPickupCommandHandler
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommandHandler
from operations.contexts.collection.application.event_handlers.drinks_ready import DrinksReadyIntegrationEventHandler
from operations.contexts.collection.domain.pickup import PickupSnapshot, Pickup
from operations.foundation.identity import derived_id
from operations.apps.runtime import command_subscription, SubscriptionWorker, complete_subscriptions
from operations.apps.composition import context_database, subscription_definitions


class CollectionContainer(containers.DeclarativeContainer):
    """Owns the shared pool; application classes depend on transport-neutral ports."""
    config = providers.Configuration()
    database = providers.Resource(context_database, "collection", config.database_url)
    pickup_transaction = providers.Singleton(PostgresAggregateTransaction[PickupSnapshot, Pickup], database, "pickup", restore_pickup_snapshot, PostgresPickupWriteRepository, pickup_publications)
    pickup_read_repository = providers.Singleton(PostgresPickupReadRepository, database)
    get_pickup = providers.Singleton(GetPickupQueryHandler, pickup_read_repository)
    list_pickups = providers.Singleton(ListPickupsQueryHandler, pickup_read_repository)
    queries = providers.Singleton(PickupQueryEndpoints, get_pickup, list_pickups)
    codec = providers.Singleton(command_codec)
    outbox = providers.Singleton(PostgresDurableCommandOutbox[OpenPickupCommand], database, codec)
    ready = providers.Singleton(DrinksReadyIntegrationEventHandler, outbox)
    open = providers.Singleton(CommandExecutor, pickup_transaction, lambda repository: OpenPickupCommandHandler(repository, derived_id))
    collect = providers.Singleton(CommandExecutor, pickup_transaction, CollectOrderCommandHandler)


def subscriptions(container: CollectionContainer) -> tuple[SubscriptionWorker, ...]:
    definitions = subscription_definitions("collection")
    return complete_subscriptions("collection", definitions, command_subscription(container.codec(), definitions, DRINKS_READY,
        lambda event: derived_id("pickup", event["orderId"]), container.ready().handle, container.open().execute))
