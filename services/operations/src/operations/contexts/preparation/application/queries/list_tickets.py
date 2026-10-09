from dataclasses import dataclass
from operations.foundation.pagination import Page, PageRequest
from operations.contexts.preparation.application.ports.ticket_reader import TicketReader
from operations.contexts.preparation.application.read_models.ticket import TicketView


@dataclass(frozen=True)
class ListTicketsQuery:
    """Select tickets or an explicitly requested page."""
    page: PageRequest | None = None


class ListTicketsQueryHandler:
    """Execute the ListTickets read use case through its owning application port."""
    def __init__(self, reader: TicketReader) -> None:
        self.reader = reader

    def execute(self, query: ListTicketsQuery) -> Page[TicketView]:
        return self.reader.page(query.page) if query.page is not None else Page(self.reader.list())
