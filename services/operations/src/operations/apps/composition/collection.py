"""Collection's providers bind its own database and durable command port."""
from enum import StrEnum
from dependency_injector import containers, providers
from operations.apps.incoming_events import DRINKS_READY
from operations.adaptors.postgres import PostgresAggregateCommandStore, PostgresAggregateQueries
from operations.adaptors.internal_commands import InternalCommandCodec, PostgresDurableCommandOutbox
from operations.adaptors.generated.cafe.internal.collection.internal_commands_pb2 import CommandEnvelope
from operations.contexts.collection.application.commands.open_pickup import OpenPickupCommand, OpenPickupCommandHandler
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommandHandler
from operations.contexts.collection.application.event_handlers import DrinksReadyIntegrationEventHandler
from operations.contexts.collection.domain import Pickup, PickupSnapshot
from operations.foundation.domain import identifier, record
from operations.foundation.identity import derived_id
from operations.apps.runtime import command_subscription, SubscriptionWorker, complete_subscriptions
from operations.apps.composition import context_database, subscription_definitions


class CollectionSubscription(StrEnum):
    """Stable queue and receipt identity, independent of Python class names."""
    OPEN_PICKUP = "collection.open-pickup"


def restore(value: object) -> PickupSnapshot:
    return Pickup.restore(value).snapshot()


def open_command(value: object) -> OpenPickupCommand:
    data = record(value)
    return OpenPickupCommand(orderId=identifier(data["orderId"]), customerId=identifier(data["customerId"]))


def command_codec() -> InternalCommandCodec[OpenPickupCommand]:
    return InternalCommandCodec("collection", CollectionSubscription.OPEN_PICKUP, "open_pickup", CommandEnvelope, open_command)


class CollectionContainer(containers.DeclarativeContainer):
    """Owns the shared pool; application classes depend on transport-neutral ports."""
    config = providers.Configuration()
    database = providers.Resource(context_database, "collection", config.database_url)
    pickups = providers.Singleton(PostgresAggregateCommandStore[PickupSnapshot], database, "pickup", restore)
    queries = providers.Singleton(PostgresAggregateQueries[PickupSnapshot], database, "pickup", restore)
    codec = providers.Singleton(command_codec)
    outbox = providers.Singleton(PostgresDurableCommandOutbox[OpenPickupCommand], database, codec)
    ready = providers.Singleton(DrinksReadyIntegrationEventHandler, outbox)
    open = providers.Singleton(OpenPickupCommandHandler, pickups, derived_id)
    collect = providers.Singleton(CollectOrderCommandHandler, pickups)


def subscriptions(container: CollectionContainer) -> tuple[SubscriptionWorker, ...]:
    definitions = subscription_definitions("collection")
    return complete_subscriptions("collection", definitions, command_subscription(container.codec(), definitions, DRINKS_READY,
        lambda event: derived_id("pickup", event["orderId"]), container.ready().handle, container.open().execute))


def command_header(body: bytes) -> tuple[str, str, str]:
    envelope = CommandEnvelope.FromString(body)
    if envelope.consumer != CollectionSubscription.OPEN_PICKUP:
        raise ValueError("Foreign command destination")
    return identifier(envelope.id), envelope.consumer+".command", identifier(envelope.correlation_id)
