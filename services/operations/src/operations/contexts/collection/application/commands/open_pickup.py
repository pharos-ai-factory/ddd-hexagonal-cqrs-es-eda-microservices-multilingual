from operations.foundation.application import RealtimeBusPort, RealtimeEvent
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.collection.application.ports.pickup_write_repository import PickupWriteRepository
from collections.abc import Callable
from typing import TypedDict
from operations.contexts.collection.domain.pickup import Pickup


class OpenPickupCommand(TypedDict):
    """Owner instruction to open collection for a prepared order."""

    orderId: str
    customerId: str


class OpenPickupCommandHandler:
    """Create one Pickup and its collection code in the command transaction."""

    def __init__(
        self,
        repository: PickupWriteRepository,
        realtime: RealtimeBusPort[RealtimeEvent],
        identities: Callable[[str, str], str],
    ) -> None:
        self.repository, self.identities = repository, identities
        self.realtime = realtime

    def execute(self, context: CommandContext, command: OpenPickupCommand) -> CommandResult:
        loaded = self.repository.get(context.target)
        state = loaded["state"] if loaded else None
        if state is not None:
            restored = state.snapshot()
            return CommandResult(restored["status"])
        code = self.identities("collection-code", command["orderId"])[:6].upper()
        pickup = Pickup.open(context.target, command["orderId"], command["customerId"], code)
        self.repository.save(pickup)
        self.realtime.publish(RealtimeEvent(snapshot=pickup.snapshot()))
        return CommandResult("ready")
