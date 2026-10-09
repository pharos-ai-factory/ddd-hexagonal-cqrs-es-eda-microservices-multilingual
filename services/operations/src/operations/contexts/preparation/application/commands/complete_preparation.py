from typing import TypedDict
from operations.contexts.preparation.domain import PreparationTicket, TicketSnapshot, DrinksReady
from operations.contracts import events
from operations.foundation.application import Change, AggregateCommandPort, Metadata, Outcome, Publication
from operations.foundation.domain import Rejection


class CompletePreparationCommand(TypedDict):
    """Requests completion of the targeted preparation ticket."""
    pass


class CompletePreparationCommandHandler:
    """Completes one ticket and records the outgoing preparation fact atomically."""
    def __init__(self, tickets: AggregateCommandPort[TicketSnapshot]) -> None:
        self.tickets = tickets

    def execute(self, metadata: Metadata, command: CompletePreparationCommand) -> Outcome:
        def decide(state: TicketSnapshot | None) -> Change[TicketSnapshot]:
            if state is None:
                raise Rejection("not_found", "The preparation ticket does not exist")
            ticket = PreparationTicket.restore(state)
            ticket.complete()
            publications = tuple(Publication("preparation.drinks-ready", events.DrinksReady(
                orderId=fact.order_id, customerId=fact.customer_id))
                for fact in ticket.events() if isinstance(fact, DrinksReady))
            return Change(ticket.snapshot(), "ready", publications=publications)
        return self.tickets.execute(metadata, decide)
