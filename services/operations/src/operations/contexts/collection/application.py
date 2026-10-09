from collections.abc import Callable
from typing import TypedDict
from operations.contexts.collection.domain import Pickup, PickupSnapshot, PickupOpened, OrderCollected
from operations.contracts import events
from operations.foundation.application import ApplicationError, Change, AggregateCommandPort, DurableCommandPort, Metadata, Outcome, Publication


class PickupNotFoundApplicationError(ApplicationError):
    """Reports the expected rejection when the targeted pickup has no stored state."""
    def __init__(self) -> None:
        super().__init__("not_found", "The pickup does not exist")


class CollectOrderCommand(TypedDict):
    """Requests collection using the supplied collection code."""
    code: str


class OpenPickupCommand(TypedDict):
    """Owner instruction to open collection for a prepared order."""
    orderId: str
    customerId: str


class DrinksReadyIntegrationEventHandler:
    """Translate Preparation's published fact into Collection's durable command."""
    def __init__(self, commands: DurableCommandPort[OpenPickupCommand]) -> None:
        self.commands = commands

    def handle(self, metadata: Metadata, event: events.DrinksReady) -> Outcome:
        return self.commands.enqueue(metadata, OpenPickupCommand(orderId=event["orderId"], customerId=event["customerId"]))


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
