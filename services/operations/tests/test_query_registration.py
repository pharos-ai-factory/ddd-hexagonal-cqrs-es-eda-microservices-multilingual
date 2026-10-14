"""Dispatch real query handlers through the production Preparation registration."""

from dependency_injector import providers
import pytest
from unittest.mock import Mock

from operations.adaptors.generated.request_names import CollectionQuery, PreparationQuery
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation import (
    preparation_requests_pb2 as wire,
)
from operations.adaptors.messaging.rabbitmq_request_registry import RabbitMQRequestRegistry
from operations.apps.composition.preparation import PreparationContainer, requests
from operations.contexts.preparation.adaptors.messaging import ticket_requests
from operations.contexts.preparation.application.read_models.ticket import TicketView
from operations.foundation.application import Loaded
from operations.foundation.pagination import Page, PageRequest

ID = "00000000-0000-4000-8000-000000000001"
MISSING = "00000000-0000-4000-8000-000000000002"
ROW: Loaded[TicketView] = {
    "exists": True,
    "version": 7,
    "state": {"id": ID, "orderId": ID, "customerId": ID, "instructions": "Coffee", "status": "queued"},
}


class Reader:
    """Observe arguments supplied to the named query handlers."""

    failed = False
    calls: list[object]

    def __init__(self) -> None:
        self.calls = []

    def get(self, identity: str) -> Loaded[TicketView] | None:
        self.calls.append(identity)
        if self.failed:
            raise RuntimeError("private database detail")
        return ROW if identity == ID else None

    def list(self) -> list[Loaded[TicketView]]:
        self.calls.append("complete")
        return [ROW] * 101

    def page(self, page: PageRequest) -> Page[TicketView]:
        self.calls.append(page)
        return Page([ROW], ID)


def query() -> wire.Request:
    return wire.Request(contract_version=1, request_id=ID, context="preparation")


def test_direct_queries_preserve_revisions_complete_lists_pagination_and_failures() -> None:
    reader = Reader()
    container = PreparationContainer()
    container.ticket_transaction.override(providers.Object(Mock()))
    container.ticket_read_repository.override(providers.Object(reader))
    registry = requests(container)
    request = query()
    request.query.get_ticket.id = ID
    reply = registry.handle(request)
    assert isinstance(reply, wire.Reply)
    assert reply.ticket.version == 7 and reply.ticket.state.id == ID
    request.query.get_ticket.id = MISSING
    assert registry.handle(request).error.code == "not_found"
    request.query.list_tickets.SetInParent()
    reply = registry.handle(request)
    assert isinstance(reply, wire.Reply)
    assert len(reply.tickets.items) == 101 and not reply.tickets.paged
    request.query.list_tickets.page.limit = 2
    request.query.list_tickets.page.after = ID
    reply = registry.handle(request)
    assert isinstance(reply, wire.Reply)
    assert reply.tickets.paged and reply.tickets.next_id == ID
    assert reader.calls[-1] == PageRequest(2, ID)
    before = len(reader.calls)
    request.query.list_tickets.page.limit = 0
    with pytest.raises(ValueError, match="page limit"):
        registry.handle(request)
    assert len(reader.calls) == before
    reader.failed = True
    request.query.get_ticket.id = ID
    reply = registry.handle(request)
    assert isinstance(reply, wire.Reply)
    assert reply.error.code == "temporarily_unavailable" and not reply.error.message
    assert reply.request_id == ID and reply.context == "preparation" and reply.contract_version == 1


def test_query_registration_rejects_duplicate_and_foreign_identifiers() -> None:
    registry = RabbitMQRequestRegistry("preparation")

    def handler(query: object) -> None:
        return None

    registry.query(
        PreparationQuery.GET_TICKET, ticket_requests.get_ticket, handler, ticket_requests.get_reply
    )
    with pytest.raises(ValueError, match="Duplicate request handler"):
        registry.query(
            PreparationQuery.GET_TICKET, ticket_requests.get_ticket, handler, ticket_requests.get_reply
        )
    with pytest.raises(ValueError, match="identifier disagrees with owner"):
        registry.query(
            CollectionQuery.GET_PICKUP, ticket_requests.get_ticket, handler, ticket_requests.get_reply
        )
