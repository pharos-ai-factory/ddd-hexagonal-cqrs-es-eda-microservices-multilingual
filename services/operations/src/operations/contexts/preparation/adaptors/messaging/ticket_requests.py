"""Explicit mapping between preparation wire contracts and application types."""

from .generated import request_payloads as inputs


from google.protobuf.message import Message
from operations.foundation.application import Loaded
from operations.contexts.preparation.application.read_models.ticket import TicketView
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_replies_pb2 import (
    Ticket,
    LoadedTicket,
    Tickets,
)
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_requests_pb2 import Reply
from operations.contracts.requests.generated.preparation import StartPreparationCommand
from operations.contracts.requests.generated.preparation import CompletePreparationCommand
from operations.contracts.requests.generated.preparation import GetTicketQuery
from operations.contracts.requests.generated.preparation import ListTicketsQuery
from operations.foundation.domain import identifier
from operations.foundation.pagination import Page


def start(value: Message) -> StartPreparationCommand:
    return inputs.start_preparation(value)


def complete(value: Message) -> CompletePreparationCommand:
    return inputs.complete_preparation(value)


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
    query = inputs.get_ticket(value)
    identifier(query.id)
    return query


def list_tickets(value: Message) -> ListTicketsQuery:
    query = inputs.list_tickets(value)
    if query.page is not None and query.page.after is not None:
        identifier(query.page.after)
    return query


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
