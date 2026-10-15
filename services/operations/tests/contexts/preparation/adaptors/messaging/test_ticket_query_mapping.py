import pytest
from operations.adaptors.persistence.mapped_read_repository import MappedReadRepository
from operations.foundation.application import Loaded
from operations.foundation.pagination import Page, PageRequest
from operations.contexts.preparation.domain.preparation_ticket import TicketSnapshot
from operations.contexts.preparation.adaptors.persistence.ticket_snapshot import restore_ticket_snapshot
from operations.contexts.preparation.adaptors.persistence.ticket_read_repository import ticket_view
from operations.contexts.preparation.adaptors.messaging import ticket_requests as wire
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_queries_pb2 import (
    GetTicket,
    ListTickets,
)
from operations.contracts.requests.generated.preparation import GetTicketQuery
from operations.contracts.requests.generated.preparation import ListTicketsQuery
from operations.contexts.preparation.application.queries.get_ticket import GetTicketQueryHandler
from operations.contexts.preparation.application.queries.list_tickets import ListTicketsQueryHandler

ID = "00000000-0000-4000-8000-000000000001"


def test_query_boundary_preserves_missing_rows_revisions_and_continuations() -> None:
    row = Loaded(
        exists=True,
        version=7,
        state=restore_ticket_snapshot(
            {"id": ID, "orderId": ID, "customerId": ID, "instructions": "Coffee", "status": "queued"}
        ),
    )
    request = PageRequest(2, ID)

    class Reader:
        def get(self, identity: str) -> Loaded[TicketSnapshot] | None:
            if identity == "failed":
                raise RuntimeError("storage unavailable")
            return row if identity == ID else None

        def list(self) -> list[Loaded[TicketSnapshot]]:
            return [row]

        def page(self, value: PageRequest) -> Page[TicketSnapshot]:
            assert value == request
            return Page([row], ID)

    reader = MappedReadRepository(Reader(), ticket_view)
    get = GetTicketQueryHandler(reader)
    listing = ListTicketsQueryHandler(reader)
    assert get.execute(GetTicketQuery("missing")) == {"error": {"code": "not_found", "message": ""}}
    assert get.execute(wire.get_ticket(GetTicket(id=ID))) == {"ticket": row}
    assert listing.execute(wire.list_tickets(ListTickets())) == {"tickets": {"items": [row], "paged": False}}
    assert listing.execute(ListTicketsQuery(request)) == {
        "tickets": {"items": [row], "paged": True, "nextId": ID}
    }
    with pytest.raises(RuntimeError, match="storage unavailable"):
        get.execute(GetTicketQuery("failed"))


def test_persistence_reader_validates_before_selecting_view_fields() -> None:
    with pytest.raises(ValueError):
        ticket_view(
            restore_ticket_snapshot(
                {"id": ID, "orderId": ID, "customerId": ID, "instructions": "Coffee", "status": "unknown"}
            )
        )
