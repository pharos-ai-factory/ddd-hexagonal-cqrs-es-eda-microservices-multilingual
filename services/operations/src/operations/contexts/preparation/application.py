from typing import TypedDict
from operations.contexts.preparation.domain import PreparationTicket, TicketState, TicketSnapshot, DrinksReady
from operations.contracts import events
from operations.foundation.application import Change, AggregateCommandPort, DurableCommandPort, Metadata, Outcome, Publication
from operations.foundation.domain import Rejection


class StartPreparationCommand(TypedDict):
    """Requests the queued-to-preparing transition for the targeted ticket."""
    pass


class CompletePreparationCommand(TypedDict):
    """Requests completion of the targeted preparation ticket."""
    pass


class AcceptOrderCommand(TypedDict):
    """Owner instruction to create a preparation ticket with immutable instructions."""
    orderId: str
    customerId: str
    instructions: str


class OrderPlacedIntegrationEventHandler:
    """Translate a published order into a durable Preparation command."""
    def __init__(self, commands: DurableCommandPort[AcceptOrderCommand]) -> None:
        self.commands = commands

    def handle(self, metadata: Metadata, event: events.OrderPlaced) -> Outcome:
        return self.commands.enqueue(metadata, AcceptOrderCommand(orderId=event["orderId"],
            customerId=event["customerId"],
            instructions="; ".join(f'{line["quantity"]} × {line["name"]}' for line in event["lines"])))


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
