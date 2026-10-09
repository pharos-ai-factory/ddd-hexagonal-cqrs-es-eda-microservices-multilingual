from dataclasses import dataclass, replace
import re
from typing import ClassVar, Literal, TypedDict
from operations.foundation.domain import CorruptState, DomainError, Rejection, identifier, record, text


class PickupAlreadyCollectedDomainError(DomainError):
    """Identifies the aggregate rule preventing a second collection."""
    def __init__(self) -> None:
        super().__init__("pickup_already_collected", "This pickup has already been collected")

type PickupStatus = Literal["ready", "collected"]


class PickupSnapshot(TypedDict):
    """Represents the persisted pickup state validated by aggregate restoration."""
    id: str
    orderId: str
    customerId: str
    code: str
    status: PickupStatus


@dataclass(frozen=True)
class PickupOpened:
    """Private fact raised when a Pickup becomes available for collection."""
    name: ClassVar[str] = "PickupOpened"
    pickup_id: str
    order_id: str
    customer_id: str
    code: str


@dataclass(frozen=True)
class OrderCollected:
    """Private fact raised when a Pickup records its single collection."""
    name: ClassVar[str] = "OrderCollected"
    order_id: str
    customer_id: str


type PickupFact = PickupOpened | OrderCollected


@dataclass(frozen=True)
class CollectionCode:
    """Validates the owner collection credential as a value object."""
    value: str

    def __post_init__(self) -> None:
        if not re.fullmatch(r"[A-Z0-9]{6}", self.value):
            raise Rejection("invalid_collection_code", "Collection codes have six uppercase letters or digits")


@dataclass(frozen=True)
class PickupState:
    """Stores private pickup lifecycle state controlled by the Pickup aggregate."""
    id: str
    order_id: str
    customer_id: str
    code: CollectionCode
    status: PickupStatus = "ready"


class Pickup:
    """Owns collection code validation and the single collection transition for an order."""
    def __init__(self, state: PickupState) -> None:
        for value in (state.id, state.order_id, state.customer_id):
            identifier(value)
        if state.status not in ("ready", "collected"):
            raise ValueError("Invalid pickup state")
        self._state = state
        self._facts: list[PickupFact] = []

    @classmethod
    def open(cls, identity: str, order: str, customer: str, code: str) -> "Pickup":
        pickup = cls(PickupState(identity, order, customer, CollectionCode(code)))
        pickup._facts.append(PickupOpened(identity, order, customer, code))
        return pickup

    @classmethod
    def restore(cls, value: object) -> "Pickup":
        try:
            fields = record(value)
            match fields["status"]:
                case "ready" | "collected" as status:
                    return cls(PickupState(identifier(fields["id"]), identifier(fields["orderId"]),
                        identifier(fields["customerId"]), CollectionCode(text(fields["code"])), status))
                case _:
                    raise ValueError("Invalid pickup status")
        except (Rejection, KeyError, TypeError, ValueError) as error:
            raise CorruptState("Corrupt pickup state") from error

    def collect(self, code: str) -> None:
        if self._state.status != "ready":
            raise PickupAlreadyCollectedDomainError()
        if code != self._state.code.value:
            raise Rejection("incorrect_collection_code", "The collection code does not match")
        self._state = replace(self._state, status="collected")
        self._facts.append(OrderCollected(self._state.order_id, self._state.customer_id))

    def snapshot(self) -> PickupSnapshot:
        s = self._state
        return {"id": s.id, "orderId": s.order_id, "customerId": s.customer_id,
                "code": s.code.value, "status": s.status}

    def events(self) -> tuple[PickupFact, ...]:
        return tuple(self._facts)
