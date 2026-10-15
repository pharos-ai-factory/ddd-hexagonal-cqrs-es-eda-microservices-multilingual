from operations.foundation.application import RealtimeBusPort, RealtimeEvent
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.preparation.application.ports.ticket_write_repository import TicketWriteRepository
from typing import TypedDict
from operations.foundation.domain import Rejection


class StartPreparationCommand(TypedDict):
    """Requests the queued-to-preparing transition for the targeted ticket."""

    pass


class StartPreparationCommandHandler:
    """Loads one ticket and applies its start rule within a command transaction."""

    def __init__(self, repository: TicketWriteRepository, realtime: RealtimeBusPort[RealtimeEvent]) -> None:
        self.repository = repository
        self.realtime = realtime

    def execute(self, context: CommandContext, command: StartPreparationCommand) -> CommandResult:
        loaded = self.repository.get(context.target)
        state = loaded["state"] if loaded else None
        if state is None:
            raise Rejection("not_found", "The preparation ticket does not exist")
        ticket = state
        ticket.start()
        self.repository.save(ticket)
        self.realtime.publish(RealtimeEvent(snapshot=ticket.snapshot()))
        return CommandResult("preparing")
