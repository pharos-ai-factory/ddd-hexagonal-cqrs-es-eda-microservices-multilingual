from operations.foundation.domain import Rejection
from operations.foundation.application import RealtimeBusPort, RealtimeEvent
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.collection.application.ports.pickup_write_repository import PickupWriteRepository
from typing import TypedDict
from operations.foundation.application import ApplicationError


class CollectOrderCommand(TypedDict):
    """Requests collection using the supplied collection code."""

    code: str


class PickupNotFoundApplicationError(ApplicationError):
    """Reports the expected rejection when the targeted pickup has no stored state."""

    def __init__(self) -> None:
        super().__init__("not_found", "The pickup does not exist")


class CollectOrderCommandHandler:
    """Checks the collection code through the aggregate and commits its collection fact."""

    def __init__(self, repository: PickupWriteRepository, realtime: RealtimeBusPort[RealtimeEvent]) -> None:
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
            self.realtime.publish(RealtimeEvent(snapshot=pickup.snapshot()))
            return CommandResult("collected")
        except Rejection as rejection:
            self.realtime.publish(RealtimeEvent(rejection=rejection.outcome()))
            raise
