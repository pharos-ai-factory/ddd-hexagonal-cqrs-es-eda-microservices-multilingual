from operations.contracts.requests.generated.preparation import ListTicketsQuery
from operations.contracts.requests.generated.preparation_responses import (
    ListTicketsResponse,
    LoadedTicket,
    Tickets,
)
from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository
from operations.foundation.pagination import Page


class ListTicketsQueryHandler:
    """Preserve revision, missing-resource and pagination policy in a generated response."""

    def __init__(self, read_repository: TicketReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: ListTicketsQuery) -> ListTicketsResponse:
        page = (
            self.read_repository.page(query.page)
            if query.page is not None
            else Page(self.read_repository.list())
        )
        result = Tickets(
            items=[
                LoadedTicket(exists=item["exists"], version=item["version"], state=item["state"])
                for item in page.items
            ],
            paged=query.page is not None,
        )
        if page.next_id is not None:
            result["nextId"] = page.next_id
        return {"tickets": result}
