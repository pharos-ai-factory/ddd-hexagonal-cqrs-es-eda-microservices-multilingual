"""Preparation's explicit providers and typed subscription declaration."""
from enum import StrEnum
from dependency_injector import containers, providers
from operations.adaptors import codec
from operations.adaptors.postgres import PostgresAggregateCommandStore, PostgresAggregateQueries
from operations.adaptors.internal_commands import InternalCommandCodec, PostgresDurableCommandOutbox
from operations.adaptors.generated.cafe.internal.preparation.internal_commands_pb2 import CommandEnvelope
from operations.contexts.preparation.application import (AcceptOrderCommand, AcceptOrderCommandHandler,
    StartPreparationCommandHandler, CompletePreparationCommandHandler, OrderPlacedIntegrationEventHandler)
from operations.contexts.preparation.domain import PreparationTicket, TicketSnapshot
from operations.foundation.domain import identifier, record, text
from operations.foundation.identity import derived_id
from operations.apps.runtime import command_subscription, SubscriptionWorker
from operations.apps.composition import context_database, subscription_definitions


class PreparationSubscription(StrEnum):
    """Stable subscription identity shared by receipt, queue and replay."""
    ACCEPT_ORDER = "preparation.accept-order"


def restore(value: object) -> TicketSnapshot:
    return PreparationTicket.restore(value).snapshot()


def accept_command(value: object) -> AcceptOrderCommand:
    data = record(value)
    return AcceptOrderCommand(orderId=identifier(data["orderId"]), customerId=identifier(data["customerId"]),
                              instructions=text(data["instructions"]))


def command_codec() -> InternalCommandCodec[AcceptOrderCommand]:
    return InternalCommandCodec("preparation", PreparationSubscription.ACCEPT_ORDER, "accept_order", CommandEnvelope, accept_command)


class PreparationContainer(containers.DeclarativeContainer):
    """Resolve the graph before starting threads; one owner resource and plain handlers."""
    config = providers.Configuration()
    database = providers.Resource(context_database, "preparation", config.database_url)
    tickets = providers.Singleton(PostgresAggregateCommandStore[TicketSnapshot], database, "ticket", restore)
    queries = providers.Singleton(PostgresAggregateQueries[TicketSnapshot], database, "ticket", restore)
    codec = providers.Singleton(command_codec)
    outbox = providers.Singleton(PostgresDurableCommandOutbox[AcceptOrderCommand], database, codec)
    accepted = providers.Singleton(OrderPlacedIntegrationEventHandler, outbox)
    accept = providers.Singleton(AcceptOrderCommandHandler, tickets)
    start = providers.Singleton(StartPreparationCommandHandler, tickets)
    complete = providers.Singleton(CompletePreparationCommandHandler, tickets)


def subscriptions(container: PreparationContainer) -> tuple[SubscriptionWorker, ...]:
    return command_subscription(container.codec(), subscription_definitions("preparation"), codec.order_placed,
        lambda event: derived_id("ticket", event["orderId"]), container.accepted().handle, container.accept().execute)


def command_header(body: bytes) -> tuple[str, str, str]:
    envelope = CommandEnvelope.FromString(body)
    if envelope.consumer != PreparationSubscription.ACCEPT_ORDER:
        raise ValueError("Foreign command destination")
    return identifier(envelope.id), envelope.consumer+".command", identifier(envelope.correlation_id)
