import pytest
from operations.adaptors.mapped_read_repository import MappedReadRepository
from operations.foundation.application import Loaded
from operations.foundation.pagination import Page, PageRequest
from operations.contexts.preparation.domain.preparation_ticket import TicketSnapshot
from operations.contexts.preparation.adaptors.persistence.ticket_snapshot import restore_ticket_snapshot
from operations.contexts.preparation.adaptors.persistence.ticket_read_repository import ticket_view
from operations.contexts.preparation.adaptors.query_endpoints import TicketQueryEndpoints
from operations.contexts.preparation.application.queries.get_ticket import GetTicketQueryHandler
from operations.contexts.preparation.application.queries.list_tickets import ListTicketsQueryHandler

ID = "00000000-0000-4000-8000-000000000001"


def test_query_boundary_preserves_missing_rows_revisions_and_continuations() -> None:
    row = Loaded(exists=True, version=7, state=restore_ticket_snapshot({"id": ID, "orderId": ID,
        "customerId": ID, "instructions": "Coffee", "status": "queued"}))
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
    endpoints = TicketQueryEndpoints(GetTicketQueryHandler(reader), ListTicketsQueryHandler(reader))
    assert endpoints.get("missing") is None
    assert endpoints.get(ID) == row
    assert endpoints.list() == [row]
    assert endpoints.page(request) == Page([row], ID)
    with pytest.raises(RuntimeError, match="storage unavailable"):
        endpoints.get("failed")


def test_persistence_reader_validates_before_selecting_view_fields() -> None:
    with pytest.raises(ValueError):
        ticket_view(restore_ticket_snapshot({"id": ID, "orderId": ID, "customerId": ID,
                      "instructions": "Coffee", "status": "unknown"}))
