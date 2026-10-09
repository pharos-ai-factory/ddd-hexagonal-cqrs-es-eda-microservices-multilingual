"""Preparation's explicit providers and typed subscription declaration."""
from operations.contexts.preparation.adaptors.persistence.ticket_write_repository import PostgresTicketWriteRepository
from operations.contexts.preparation.adaptors.persistence.ticket_snapshot import restore_ticket_snapshot
from operations.contexts.preparation.adaptors.persistence.ticket_read_repository import PostgresTicketReadRepository
from operations.contexts.preparation.adaptors.messaging.accept_order_codec import command_codec
from operations.contexts.preparation.adaptors.query_endpoints import TicketQueryEndpoints
from operations.contexts.preparation.application.queries.get_ticket import GetTicketQueryHandler
from operations.contexts.preparation.application.queries.list_tickets import ListTicketsQueryHandler
from operations.contexts.preparation.adaptors.messaging.ticket_publications import ticket_publications
from dependency_injector import containers, providers
from operations.adaptors.command_execution import CommandExecutor
from operations.contexts.preparation.adaptors.messaging.incoming_event import ORDER_PLACED
from operations.adaptors.aggregate_transaction import PostgresAggregateTransaction
from operations.adaptors.internal_commands import PostgresDurableCommandOutbox
from operations.contexts.preparation.application.commands.accept_order import AcceptOrderCommand, AcceptOrderCommandHandler
from operations.contexts.preparation.application.commands.start_preparation import StartPreparationCommandHandler
from operations.contexts.preparation.application.commands.complete_preparation import CompletePreparationCommandHandler
from operations.contexts.preparation.application.event_handlers.order_placed import OrderPlacedIntegrationEventHandler
from operations.contexts.preparation.domain.preparation_ticket import TicketSnapshot, PreparationTicket
from operations.foundation.identity import derived_id
from operations.apps.runtime import command_subscription, SubscriptionWorker, complete_subscriptions
from operations.apps.composition import context_database, subscription_definitions


class PreparationContainer(containers.DeclarativeContainer):
    """Resolve the graph before starting threads; one owner resource and plain handlers."""
    config = providers.Configuration()
    database = providers.Resource(context_database, "preparation", config.database_url)
    ticket_transaction = providers.Singleton(PostgresAggregateTransaction[TicketSnapshot, PreparationTicket], database, "ticket", restore_ticket_snapshot, PostgresTicketWriteRepository, ticket_publications)
    ticket_read_repository = providers.Singleton(PostgresTicketReadRepository, database)
    get_ticket = providers.Singleton(GetTicketQueryHandler, ticket_read_repository)
    list_tickets = providers.Singleton(ListTicketsQueryHandler, ticket_read_repository)
    queries = providers.Singleton(TicketQueryEndpoints, get_ticket, list_tickets)
    codec = providers.Singleton(command_codec)
    outbox = providers.Singleton(PostgresDurableCommandOutbox[AcceptOrderCommand], database, codec)
    accepted = providers.Singleton(OrderPlacedIntegrationEventHandler, outbox)
    accept = providers.Singleton(CommandExecutor, ticket_transaction, AcceptOrderCommandHandler)
    start = providers.Singleton(CommandExecutor, ticket_transaction, StartPreparationCommandHandler)
    complete = providers.Singleton(CommandExecutor, ticket_transaction, CompletePreparationCommandHandler)


def subscriptions(container: PreparationContainer) -> tuple[SubscriptionWorker, ...]:
    definitions = subscription_definitions("preparation")
    return complete_subscriptions("preparation", definitions, command_subscription(container.codec(), definitions, ORDER_PLACED,
        lambda event: derived_id("ticket", event["orderId"]), container.accepted().handle, container.accept().execute))
