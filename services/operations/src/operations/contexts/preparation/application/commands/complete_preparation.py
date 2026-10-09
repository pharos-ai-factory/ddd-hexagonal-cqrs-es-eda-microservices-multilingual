from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.preparation.application.ports.ticket_write_repository import TicketWriteRepository
from typing import TypedDict
from operations.foundation.domain import Rejection


class CompletePreparationCommand(TypedDict):
    """Requests completion of the targeted preparation ticket."""
    pass


class CompletePreparationCommandHandler:
    """Completes one ticket and records the outgoing preparation fact atomically."""
    def __init__(self, repository: TicketWriteRepository) -> None:
        self.repository = repository

    def execute(self, context: CommandContext, command: CompletePreparationCommand) -> CommandResult:
        loaded = self.repository.get(context.target)
        state = loaded["state"] if loaded else None
        if state is None:
            raise Rejection("not_found", "The preparation ticket does not exist")
        ticket = state
        ticket.complete()
        self.repository.save(ticket)
        return CommandResult("ready")
