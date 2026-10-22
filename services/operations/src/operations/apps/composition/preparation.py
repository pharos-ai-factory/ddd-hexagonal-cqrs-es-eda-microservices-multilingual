"""Preparation's explicit providers and typed subscription declaration."""

from operations.contexts.preparation.adaptors.messaging.generated import response_payloads as responses

from operations.adaptors.persistence.realtime_event_bus import PostgresRealtimeMessageBus
from operations.contexts.preparation.application.event_handlers.realtime_drinks_ready import (
    DrinksReadyRealtimeDomainEventHandler,
)
from operations.apps.runtime import realtime_subscription


from operations.adaptors.generated.message_names import PreparationIntegrationEvent, PreparationSubscription

from operations.contexts.preparation.adaptors.persistence.ticket_write_repository import (
    PostgresTicketWriteRepository,
)
from operations.contexts.preparation.adaptors.persistence.ticket_snapshot import restore_ticket_snapshot
from operations.contexts.preparation.adaptors.persistence.ticket_read_repository import (
    PostgresTicketReadRepository,
)
from operations.contexts.preparation.adaptors.messaging.accept_order_codec import command_codec
from operations.adaptors.generated.request_names import PreparationCommand
from operations.contexts.preparation.application.queries.get_ticket import GetTicketQueryHandler
from operations.contexts.preparation.application.queries.list_tickets import ListTicketsQueryHandler
from operations.contexts.preparation.adaptors.messaging.ticket_publications import ticket_publications
from dependency_injector import containers, providers
from operations.apps.command_registration import CommandRegistration
from operations.adaptors.messaging.rabbitmq_request_registry import RabbitMQRequestRegistry
from operations.contexts.preparation.adaptors.messaging import ticket_requests as wire
from operations.contexts.preparation.adaptors.messaging.order_placed_subscription import ORDER_PLACED
from operations.adaptors.persistence.aggregate_transaction import PostgresAggregateTransaction
from operations.adaptors.persistence.command_outbox import PostgresDurableCommandOutbox
from operations.contexts.preparation.application.commands.accept_order import AcceptOrderCommandHandler
from operations.contracts.requests.generated.preparation import AcceptOrderCommand
from operations.contexts.preparation.application.commands.start_preparation import (
    StartPreparationCommandHandler,
)
from operations.contexts.preparation.application.commands.complete_preparation import (
    CompletePreparationCommandHandler,
)
from operations.contexts.preparation.application.event_handlers.order_placed import (
    OrderPlacedIntegrationEventHandler,
)
from operations.contexts.preparation.domain.preparation_ticket import TicketSnapshot, PreparationTicket
from operations.foundation.identity import derived_id
from operations.apps.runtime import SubscriptionWorker, complete_subscriptions
from operations.apps.composition import context_database, subscription_definitions

from operations.adaptors.persistence.integration_event_outbox import PostgresPublicationOutbox
from operations.apps.runtime import publication_subscription
from operations.adaptors.generated.event_preparation import publish_DrinksReady
from operations.contexts.preparation.application.event_handlers.publish_drinks_ready import (
    DrinksReadyDomainEventHandler,
)
from operations.contexts.preparation.adaptors.messaging.private_events import DRINKS_READY_DOMAIN


class PreparationContainer(containers.DeclarativeContainer):
    """Resolve the graph before starting threads; one owner resource and plain handlers."""

    config = providers.Configuration()
    database = providers.Resource(context_database, "preparation", config.database_url)
    ticket_transaction = providers.Singleton(
        PostgresAggregateTransaction[TicketSnapshot, PreparationTicket],
        database,
        "ticket",
        restore_ticket_snapshot,
        PostgresTicketWriteRepository,
        ticket_publications,
    )
    ticket_read_repository = providers.Singleton(PostgresTicketReadRepository, database)
    get_ticket = providers.Singleton(GetTicketQueryHandler, ticket_read_repository)
    list_tickets = providers.Singleton(ListTicketsQueryHandler, ticket_read_repository)
    codec = providers.Singleton(command_codec)
    outbox = providers.Singleton(PostgresDurableCommandOutbox[AcceptOrderCommand], database, codec)
    accepted = providers.Singleton(OrderPlacedIntegrationEventHandler, outbox)
    commands = providers.Singleton(CommandRegistration[PreparationTicket], ticket_transaction)

    drinks_ready_publications = providers.Singleton(
        PostgresPublicationOutbox,
        database,
        PreparationSubscription.PUBLISH_DRINKS_READY,
        "ticket",
        publish_DrinksReady,
    )
    realtime_bus = providers.Singleton(
        PostgresRealtimeMessageBus, database, PreparationSubscription.REALTIME_DRINKS_READY, "ticket"
    )
    realtime_drinks_ready = providers.Singleton(
        DrinksReadyRealtimeDomainEventHandler, ticket_read_repository, realtime_bus
    )
    publish_drinks_ready = providers.Singleton(DrinksReadyDomainEventHandler, drinks_ready_publications)


def requests(container: PreparationContainer) -> RabbitMQRequestRegistry:
    """Map published preparation commands and queries directly to their handlers."""
    registry = RabbitMQRequestRegistry("preparation")
    commands = container.commands()
    commands.request(
        registry, PreparationCommand.START_PREPARATION, wire.start, StartPreparationCommandHandler
    )
    commands.request(
        registry,
        PreparationCommand.COMPLETE_PREPARATION,
        wire.complete,
        lambda repository, _realtime: CompletePreparationCommandHandler(repository),
    )
    responses.register_get_ticket(registry, wire.get_ticket, container.get_ticket().execute)
    responses.register_list_tickets(registry, wire.list_tickets, container.list_tickets().execute)
    return registry


def subscriptions(container: PreparationContainer) -> tuple[SubscriptionWorker, ...]:
    definitions = subscription_definitions("preparation")
    return complete_subscriptions(
        "preparation",
        definitions,
        container.commands().subscription(
            container.codec(),
            definitions,
            ORDER_PLACED,
            lambda event: derived_id("ticket", event["orderId"]),
            container.accepted().handle,
            AcceptOrderCommandHandler,
        )
        + publication_subscription(
            "preparation",
            PreparationSubscription.PUBLISH_DRINKS_READY,
            definitions,
            DRINKS_READY_DOMAIN,
            PreparationIntegrationEvent.DRINKS_READY,
            container.publish_drinks_ready().handle,
        )
        + realtime_subscription(
            "preparation",
            PreparationSubscription.REALTIME_DRINKS_READY,
            definitions,
            DRINKS_READY_DOMAIN,
            "ticket",
            container.realtime_drinks_ready().handle,
        ),
    )
