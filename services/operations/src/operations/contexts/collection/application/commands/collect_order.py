from typing import TypedDict
from operations.contexts.collection.domain import Pickup, PickupSnapshot, OrderCollected
from operations.contracts import events
from operations.foundation.application import ApplicationError, Change, AggregateCommandPort, Metadata, Outcome, Publication


class CollectOrderCommand(TypedDict):
    """Requests collection using the supplied collection code."""
    code: str


class PickupNotFoundApplicationError(ApplicationError):
    """Reports the expected rejection when the targeted pickup has no stored state."""
    def __init__(self) -> None:
        super().__init__("not_found", "The pickup does not exist")


class CollectOrderCommandHandler:
    """Checks the collection code through the aggregate and commits its collection fact."""
    def __init__(self, pickups: AggregateCommandPort[PickupSnapshot]) -> None:
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
