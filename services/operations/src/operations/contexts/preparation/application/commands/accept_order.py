from operations.contracts.requests.generated.preparation import AcceptOrderCommand
from operations.foundation.application import (
    RealtimeBusPort,
    RealtimeMessage,
    context_target,
    snapshot_message,
)
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.preparation.application.ports.ticket_write_repository import TicketWriteRepository
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketState


class AcceptOrderCommandHandler:
    """Create one PreparationTicket; redelivery preserves the existing ticket."""

    def __init__(self, repository: TicketWriteRepository, realtime: RealtimeBusPort[RealtimeMessage]) -> None:
        self.repository = repository
        self.realtime = realtime

    def execute(self, context: CommandContext, command: AcceptOrderCommand) -> CommandResult:
        loaded = self.repository.get(context.target)
        state = loaded["state"] if loaded else None
        if state is not None:
            restored = state.snapshot()
            return CommandResult(restored["status"])
        ticket = PreparationTicket(
            TicketState(context.target, command["orderId"], command["customerId"], command["instructions"])
        )
        self.repository.save(ticket)
        self.realtime.publish(snapshot_message(context_target("preparation"), ticket.snapshot()))
        return CommandResult("queued")
