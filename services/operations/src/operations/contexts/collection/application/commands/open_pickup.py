from collections.abc import Callable
from typing import TypedDict
from operations.contexts.collection.domain import Pickup, PickupSnapshot, PickupOpened
from operations.contracts import events
from operations.foundation.application import Change, AggregateCommandPort, Metadata, Outcome, Publication


class OpenPickupCommand(TypedDict):
    """Owner instruction to open collection for a prepared order."""
    orderId: str
    customerId: str


class OpenPickupCommandHandler:
    """Create one Pickup and its collection code in the command transaction."""
    def __init__(self, pickups: AggregateCommandPort[PickupSnapshot], identities: Callable[[str, str], str]) -> None:
        self.pickups, self.identities = pickups, identities

    def execute(self, metadata: Metadata, command: OpenPickupCommand) -> Outcome:
        def decide(state: PickupSnapshot | None) -> Change[PickupSnapshot]:
            if state is not None:
                restored = Pickup.restore(state).snapshot()
                return Change(restored, restored["status"], changed=False)
            code = self.identities("collection-code", command["orderId"])[:6].upper()
            pickup = Pickup.open(metadata.target, command["orderId"], command["customerId"], code)
            publications = tuple(Publication("collection.pickup-opened", events.PickupOpened(
                pickupId=fact.pickup_id, orderId=fact.order_id, customerId=fact.customer_id, collectionCode=fact.code))
                for fact in pickup.events() if isinstance(fact, PickupOpened))
            return Change(pickup.snapshot(), "ready", publications=publications)
        return self.pickups.execute(metadata, decide)
