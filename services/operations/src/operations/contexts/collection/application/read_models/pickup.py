from typing import Literal, TypedDict


class PickupView(TypedDict):
    """Application read contract for one pickup, independent of storage and transport."""
    id: str
    orderId: str
    customerId: str
    code: str
    status: Literal["ready", "collected"]
