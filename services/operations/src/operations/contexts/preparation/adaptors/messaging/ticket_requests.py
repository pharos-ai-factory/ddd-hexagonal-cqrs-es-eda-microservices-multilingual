"""Explicit mapping between preparation wire contracts and application types."""

from .generated import request_payloads as inputs


from google.protobuf.message import Message
from operations.contracts.requests.generated.preparation import StartPreparationCommand
from operations.contracts.requests.generated.preparation import CompletePreparationCommand
from operations.contracts.requests.generated.preparation import GetTicketQuery
from operations.contracts.requests.generated.preparation import ListTicketsQuery
from operations.foundation.domain import identifier


def start(value: Message) -> StartPreparationCommand:
    return inputs.start_preparation(value)


def complete(value: Message) -> CompletePreparationCommand:
    return inputs.complete_preparation(value)


def get_ticket(value: Message) -> GetTicketQuery:
    query = inputs.get_ticket(value)
    identifier(query.id)
    return query


def list_tickets(value: Message) -> ListTicketsQuery:
    query = inputs.list_tickets(value)
    if query.page is not None and query.page.after is not None:
        identifier(query.page.after)
    return query
