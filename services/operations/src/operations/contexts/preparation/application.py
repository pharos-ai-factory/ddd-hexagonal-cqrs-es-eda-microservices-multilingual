from typing import TypedDict
from operations.contexts.preparation.domain import PreparationTicket, TicketState, TicketSnapshot, DrinksReady
from operations.contracts import events
from operations.foundation.application import Change, CommandPort, Metadata, Outcome, Publication
from operations.foundation.domain import Rejection


class StartPreparationCommand(TypedDict):
    pass


class CompletePreparationCommand(TypedDict):
    pass


class AcceptOrder:
    def __init__(self, tickets: CommandPort[TicketSnapshot]) -> None:
        self.tickets = tickets

    def handle(self, metadata: Metadata, event: events.OrderPlaced) -> Outcome:
        def decide(state: TicketSnapshot | None) -> Change[TicketSnapshot]:
            if state is not None:
                restored = PreparationTicket.restore(state).snapshot()
                return Change(restored, restored["status"], changed=False)
            instructions = "; ".join(f'{line["quantity"]} × {line["name"]}' for line in event["lines"])
            ticket = PreparationTicket(TicketState(
                metadata.target, event["orderId"], event["customerId"], instructions))
            return Change(ticket.snapshot(), "queued")
        return self.tickets.execute(metadata, decide)


class StartPreparation:
    def __init__(self, tickets: CommandPort[TicketSnapshot]) -> None:
        self.tickets = tickets

    def execute(self, metadata: Metadata, command: StartPreparationCommand) -> Outcome:
        def decide(state: TicketSnapshot | None) -> Change[TicketSnapshot]:
            if state is None:
                raise Rejection("not_found", "The preparation ticket does not exist")
            ticket = PreparationTicket.restore(state)
            ticket.start()
            return Change(ticket.snapshot(), "preparing")
        return self.tickets.execute(metadata, decide)


class CompletePreparation:
    def __init__(self, tickets: CommandPort[TicketSnapshot]) -> None:
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
