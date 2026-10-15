from operations.contracts.requests.generated.preparation import ListTicketsQuery
from operations.foundation.pagination import Page
from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository
from operations.contexts.preparation.application.read_models.ticket import TicketView


class ListTicketsQueryHandler:
    """Execute the ListTickets read use case through its owning application port."""

    def __init__(self, read_repository: TicketReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: ListTicketsQuery) -> Page[TicketView]:
        return (
            self.read_repository.page(query.page)
            if query.page is not None
            else Page(self.read_repository.list())
        )
