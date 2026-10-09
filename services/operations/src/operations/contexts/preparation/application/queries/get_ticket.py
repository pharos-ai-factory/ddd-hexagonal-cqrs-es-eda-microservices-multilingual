from dataclasses import dataclass
from operations.foundation.application import Loaded
from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository
from operations.contexts.preparation.application.read_models.ticket import TicketView


@dataclass(frozen=True)
class GetTicketQuery:
    """Select one ticket."""
    identity: str


class GetTicketQueryHandler:
    """Execute the GetTicket read use case through its owning application port."""
    def __init__(self, read_repository: TicketReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: GetTicketQuery) -> Loaded[TicketView] | None:
        return self.read_repository.get(query.identity)
