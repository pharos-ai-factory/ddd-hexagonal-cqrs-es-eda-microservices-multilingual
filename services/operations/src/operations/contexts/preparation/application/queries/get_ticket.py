from dataclasses import dataclass
from operations.foundation.application import Loaded
from operations.contexts.preparation.application.ports.ticket_reader import TicketReader
from operations.contexts.preparation.application.read_models.ticket import TicketView


@dataclass(frozen=True)
class GetTicketQuery:
    """Select one ticket."""
    identity: str


class GetTicketQueryHandler:
    """Execute the GetTicket read use case through its owning application port."""
    def __init__(self, reader: TicketReader) -> None:
        self.reader = reader

    def execute(self, query: GetTicketQuery) -> Loaded[TicketView] | None:
        return self.reader.get(query.identity)
