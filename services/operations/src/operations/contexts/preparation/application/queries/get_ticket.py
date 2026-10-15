from operations.contracts.requests.generated.preparation import GetTicketQuery
from operations.foundation.application import Loaded
from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository
from operations.contexts.preparation.application.read_models.ticket import TicketView


class GetTicketQueryHandler:
    """Execute the GetTicket read use case through its owning application port."""

    def __init__(self, read_repository: TicketReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: GetTicketQuery) -> Loaded[TicketView] | None:
        return self.read_repository.get(query.id)
