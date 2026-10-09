from dataclasses import dataclass
from operations.foundation.pagination import Page, PageRequest
from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository
from operations.contexts.preparation.application.read_models.ticket import TicketView


@dataclass(frozen=True)
class ListTicketsQuery:
    """Select tickets or an explicitly requested page."""
    page: PageRequest | None = None


class ListTicketsQueryHandler:
    """Execute the ListTickets read use case through its owning application port."""
    def __init__(self, read_repository: TicketReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: ListTicketsQuery) -> Page[TicketView]:
        return self.read_repository.page(query.page) if query.page is not None else Page(self.read_repository.list())
