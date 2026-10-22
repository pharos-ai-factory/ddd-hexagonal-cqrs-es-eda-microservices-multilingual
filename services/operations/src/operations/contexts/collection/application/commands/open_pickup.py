from operations.contracts.requests.generated.collection import OpenPickupCommand
from operations.foundation.application import (
    RealtimeBusPort,
    RealtimeMessage,
    context_target,
    snapshot_message,
)
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.collection.application.ports.pickup_write_repository import PickupWriteRepository
from collections.abc import Callable
from operations.contexts.collection.domain.pickup import Pickup


class OpenPickupCommandHandler:
    """Create one Pickup and its collection code in the command transaction."""

    def __init__(
        self,
        repository: PickupWriteRepository,
        realtime: RealtimeBusPort[RealtimeMessage],
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
        self.realtime.publish(snapshot_message(context_target("collection"), pickup.snapshot()))
        return CommandResult("ready")
