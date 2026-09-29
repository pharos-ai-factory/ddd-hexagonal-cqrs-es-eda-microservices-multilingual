from collections.abc import Callable
from typing import TypedDict
from operations.contexts.collection.domain import Pickup, PickupSnapshot, PickupOpened, OrderCollected
from operations.contracts import events
from operations.foundation.application import ApplicationError, Change, CommandPort, Metadata, Outcome, Publication


class PickupNotFoundApplicationError(ApplicationError):
    def __init__(self) -> None:
        super().__init__("not_found", "The pickup does not exist")


class CollectOrderCommand(TypedDict):
    code: str


class OpenPickup:
    def __init__(self, pickups: CommandPort[PickupSnapshot], identities: Callable[[str, str], str]) -> None:
        self.pickups, self.identities = pickups, identities

    def handle(self, metadata: Metadata, event: events.DrinksReady) -> Outcome:
        def decide(state: PickupSnapshot | None) -> Change[PickupSnapshot]:
            if state is not None:
                restored = Pickup.restore(state).snapshot()
                return Change(restored, restored["status"], changed=False)
            code = self.identities("collection-code", event["orderId"])[:6].upper()
            pickup = Pickup.open(metadata.target, event["orderId"], event["customerId"], code)
            publications = tuple(Publication("collection.pickup-opened", events.PickupOpened(
                pickupId=fact.pickup_id, orderId=fact.order_id, customerId=fact.customer_id, collectionCode=fact.code))
                for fact in pickup.events() if isinstance(fact, PickupOpened))
            return Change(pickup.snapshot(), "ready", publications=publications)
        return self.pickups.execute(metadata, decide)


class CollectOrder:
    def __init__(self, pickups: CommandPort[PickupSnapshot]) -> None:
        self.pickups = pickups

    def execute(self, metadata: Metadata, command: CollectOrderCommand) -> Outcome:
        def decide(state: PickupSnapshot | None) -> Change[PickupSnapshot]:
            if state is None:
                raise PickupNotFoundApplicationError()
            pickup = Pickup.restore(state)
            pickup.collect(command["code"])
            publications = tuple(Publication("collection.order-collected", events.OrderCollected(
                orderId=fact.order_id, customerId=fact.customer_id))
                for fact in pickup.events() if isinstance(fact, OrderCollected))
            return Change(pickup.snapshot(), "collected", publications=publications)
        return self.pickups.execute(metadata, decide)
