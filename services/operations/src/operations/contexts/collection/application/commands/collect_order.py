from operations.contracts.requests.generated.collection import CollectOrderCommand
from operations.foundation.domain import Rejection
from operations.foundation.application import (
    RealtimeBusPort,
    RealtimeMessage,
    context_target,
    principal_target,
    rejection_message,
    snapshot_message,
)
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.collection.application.ports.pickup_write_repository import PickupWriteRepository
from operations.foundation.application import ApplicationError


class PickupNotFoundApplicationError(ApplicationError):
    """Reports the expected rejection when the targeted pickup has no stored state."""

    def __init__(self) -> None:
        super().__init__("not_found", "The pickup does not exist")


class CollectOrderCommandHandler:
    """Checks the collection code through the aggregate and commits its collection fact."""

    def __init__(self, repository: PickupWriteRepository, realtime: RealtimeBusPort[RealtimeMessage]) -> None:
        self.repository = repository
        self.realtime = realtime

    def execute(self, context: CommandContext, command: CollectOrderCommand) -> CommandResult:
        try:
            loaded = self.repository.get(context.target)
            state = loaded["state"] if loaded else None
            if state is None:
                raise PickupNotFoundApplicationError()
            pickup = state
            pickup.collect(command["code"])
            self.repository.save(pickup)
            self.realtime.publish(snapshot_message(context_target("collection"), pickup.snapshot()))
            return CommandResult("collected")
        except Rejection as rejection:
            if context.actor_id:
                self.realtime.publish(
                    rejection_message(principal_target(context.actor_id), rejection.outcome())
                )
            raise
