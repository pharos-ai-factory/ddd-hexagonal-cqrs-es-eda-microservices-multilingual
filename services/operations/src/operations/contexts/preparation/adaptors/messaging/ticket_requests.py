"""Explicit mapping between preparation wire contracts and application types."""

from google.protobuf.message import Message
from operations.foundation.application import Loaded
from operations.contexts.preparation.application.read_models.ticket import TicketView
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_replies_pb2 import (
    Ticket,
    LoadedTicket,
    Tickets,
)
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_requests_pb2 import Reply
from operations.contexts.preparation.application.commands.start_preparation import StartPreparationCommand
from operations.contexts.preparation.application.commands.complete_preparation import (
    CompletePreparationCommand,
)
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_commands_pb2 import (
    StartPreparation,
    CompletePreparation,
)
from operations.contexts.preparation.application.queries.get_ticket import GetTicketQuery
from operations.contexts.preparation.application.queries.list_tickets import ListTicketsQuery
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_queries_pb2 import (
    GetTicket,
    ListTickets,
)
from operations.foundation.domain import identifier
from operations.foundation.pagination import Page, PageRequest


def start(value: Message) -> StartPreparationCommand:
    if not isinstance(value, StartPreparation):
        raise ValueError("Unexpected start preparation payload")
    return StartPreparationCommand()


def complete(value: Message) -> CompletePreparationCommand:
    if not isinstance(value, CompletePreparation):
        raise ValueError("Unexpected complete preparation payload")
    return CompletePreparationCommand()


def item(loaded: Loaded[TicketView]) -> LoadedTicket:
    s = loaded["state"]
    return LoadedTicket(
        exists=loaded["exists"],
        version=loaded["version"],
        state=Ticket(
            id=s["id"],
            order_id=s["orderId"],
            customer_id=s["customerId"],
            instructions=s["instructions"],
            status=s["status"],
        ),
    )


def listing(items: list[Loaded[TicketView]], paged: bool, next_id: str | None) -> Reply:
    result = Tickets(items=[item(value) for value in items], paged=paged)
    if next_id is not None:
        result.next_id = next_id
    return Reply(tickets=result)


def get_ticket(value: Message) -> GetTicketQuery:
    """Decode the published item query into its application input."""
    if not isinstance(value, GetTicket):
        raise ValueError("Unexpected get ticket payload")
    return GetTicketQuery(identifier(value.id))


def list_tickets(value: Message) -> ListTicketsQuery:
    """Preserve complete reads and explicitly requested pagination."""
    if not isinstance(value, ListTickets):
        raise ValueError("Unexpected list tickets payload")
    page = value.page if value.HasField("page") else None
    return ListTicketsQuery(
        PageRequest(page.limit, identifier(page.after) if page.HasField("after") else None)
        if page is not None
        else None
    )


def get_reply(query: GetTicketQuery, loaded: Loaded[TicketView] | None) -> Reply:
    """Map a missing resource or its current revision to the owner reply."""
    if loaded is None:
        reply = Reply()
        reply.error.code = "not_found"
        return reply
    return Reply(ticket=item(loaded))


def list_reply(query: ListTicketsQuery, page: Page[TicketView]) -> Reply:
    """Retain the pagination flag and continuation from the application result."""
    return listing(page.items, query.page is not None, page.next_id)
