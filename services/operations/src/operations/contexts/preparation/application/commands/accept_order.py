from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.preparation.application.ports.ticket_write_repository import TicketWriteRepository
from typing import TypedDict
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketState


class AcceptOrderCommand(TypedDict):
    """Owner instruction to create a preparation ticket with immutable instructions."""
    orderId: str
    customerId: str
    instructions: str


class AcceptOrderCommandHandler:
    """Create one PreparationTicket; redelivery preserves the existing ticket."""
    def __init__(self, repository: TicketWriteRepository) -> None:
        self.repository = repository

    def execute(self, context: CommandContext, command: AcceptOrderCommand) -> CommandResult:
        loaded = self.repository.get(context.target)
        state = loaded["state"] if loaded else None
        if state is not None:
            restored = state.snapshot()
            return CommandResult(restored["status"])
        ticket = PreparationTicket(TicketState(
            context.target, command["orderId"], command["customerId"], command["instructions"]))
        self.repository.save(ticket)
        return CommandResult("queued")
