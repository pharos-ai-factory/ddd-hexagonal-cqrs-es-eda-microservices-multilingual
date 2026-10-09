from typing import TypedDict
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketSnapshot
from operations.foundation.application import Change, AggregateCommandPort, Metadata, Outcome
from operations.foundation.domain import Rejection


class StartPreparationCommand(TypedDict):
    """Requests the queued-to-preparing transition for the targeted ticket."""
    pass


class StartPreparationCommandHandler:
    """Loads one ticket and applies its start rule within a command transaction."""
    def __init__(self, tickets: AggregateCommandPort[TicketSnapshot]) -> None:
        self.tickets = tickets

    def execute(self, metadata: Metadata, command: StartPreparationCommand) -> Outcome:
        def decide(state: TicketSnapshot | None) -> Change[TicketSnapshot]:
            if state is None:
                raise Rejection("not_found", "The preparation ticket does not exist")
            ticket = PreparationTicket.restore(state)
            ticket.start()
            return Change(ticket.snapshot(), "preparing")
        return self.tickets.execute(metadata, decide)
