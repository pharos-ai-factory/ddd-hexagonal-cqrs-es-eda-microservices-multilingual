from typing import TypedDict
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketState, TicketSnapshot
from operations.foundation.application import Change, AggregateCommandPort, Metadata, Outcome


class AcceptOrderCommand(TypedDict):
    """Owner instruction to create a preparation ticket with immutable instructions."""
    orderId: str
    customerId: str
    instructions: str


class AcceptOrderCommandHandler:
    """Create one PreparationTicket; redelivery preserves the existing ticket."""
    def __init__(self, tickets: AggregateCommandPort[TicketSnapshot]) -> None:
        self.tickets = tickets

    def execute(self, metadata: Metadata, command: AcceptOrderCommand) -> Outcome:
        def decide(state: TicketSnapshot | None) -> Change[TicketSnapshot]:
            if state is not None:
                restored = PreparationTicket.restore(state).snapshot()
                return Change(restored, restored["status"], changed=False)
            ticket = PreparationTicket(TicketState(
                metadata.target, command["orderId"], command["customerId"], command["instructions"]))
            return Change(ticket.snapshot(), "queued")
        return self.tickets.execute(metadata, decide)
