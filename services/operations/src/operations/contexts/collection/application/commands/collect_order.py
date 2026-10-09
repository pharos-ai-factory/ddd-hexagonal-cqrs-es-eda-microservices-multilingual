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
    def __init__(self, repository: PickupWriteRepository) -> None:
        self.repository = repository

    def execute(self, context: CommandContext, command: CollectOrderCommand) -> CommandResult:
        loaded = self.repository.get(context.target)
        state = loaded["state"] if loaded else None
        if state is None:
            raise PickupNotFoundApplicationError()
        pickup = state
        pickup.collect(command["code"])
        self.repository.save(pickup)
        return CommandResult("collected")
