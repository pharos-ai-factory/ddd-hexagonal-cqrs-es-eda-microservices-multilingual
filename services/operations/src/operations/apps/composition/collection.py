"""Collection's providers bind its own database and durable command port."""
from operations.contexts.collection.adaptors.persistence.pickups import restore, PostgresPickupReader
from operations.contexts.collection.adaptors.messaging.open_pickup_codec import command_codec
from operations.contexts.collection.adaptors.query_endpoints import PickupQueryEndpoints
from operations.contexts.collection.application.queries.get_pickup import GetPickupQueryHandler
from operations.contexts.collection.application.queries.list_pickups import ListPickupsQueryHandler
from dependency_injector import containers, providers
from operations.contexts.collection.adaptors.messaging.incoming_event import DRINKS_READY
from operations.adaptors.postgres import PostgresAggregateCommandStore
from operations.adaptors.internal_commands import PostgresDurableCommandOutbox
from operations.contexts.collection.application.commands.open_pickup import OpenPickupCommand, OpenPickupCommandHandler
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommandHandler
from operations.contexts.collection.application.event_handlers.drinks_ready import DrinksReadyIntegrationEventHandler
from operations.contexts.collection.domain.pickup import PickupSnapshot
from operations.foundation.identity import derived_id
from operations.apps.runtime import command_subscription, SubscriptionWorker, complete_subscriptions
from operations.apps.composition import context_database, subscription_definitions


class CollectionContainer(containers.DeclarativeContainer):
    """Owns the shared pool; application classes depend on transport-neutral ports."""
    config = providers.Configuration()
    database = providers.Resource(context_database, "collection", config.database_url)
    pickups = providers.Singleton(PostgresAggregateCommandStore[PickupSnapshot], database, "pickup", restore)
    reader = providers.Singleton(PostgresPickupReader, database)
    get_pickup = providers.Singleton(GetPickupQueryHandler, reader)
    list_pickups = providers.Singleton(ListPickupsQueryHandler, reader)
    queries = providers.Singleton(PickupQueryEndpoints, get_pickup, list_pickups)
    codec = providers.Singleton(command_codec)
    outbox = providers.Singleton(PostgresDurableCommandOutbox[OpenPickupCommand], database, codec)
    ready = providers.Singleton(DrinksReadyIntegrationEventHandler, outbox)
    open = providers.Singleton(OpenPickupCommandHandler, pickups, derived_id)
    collect = providers.Singleton(CollectOrderCommandHandler, pickups)


def subscriptions(container: CollectionContainer) -> tuple[SubscriptionWorker, ...]:
    definitions = subscription_definitions("collection")
    return complete_subscriptions("collection", definitions, command_subscription(container.codec(), definitions, DRINKS_READY,
        lambda event: derived_id("pickup", event["orderId"]), container.ready().handle, container.open().execute))
