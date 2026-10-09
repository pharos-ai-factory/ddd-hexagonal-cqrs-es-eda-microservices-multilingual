"""The published event shapes consumed or emitted by Operations."""
from typing import TypedDict


class OrderLine(TypedDict):
    """Carries one immutable line from a published order."""
    id: str
    offerCode: str
    name: str
    quantity: int
    minor: int


class OrderPlaced(TypedDict):
    """Carries the published immutable ordering snapshot consumed by Preparation."""
    orderId: str
    customerId: str
    editionId: str
    currency: str
    lines: list[OrderLine]


class DrinksReady(TypedDict):
    """Published Preparation fact consumed by Collection to request a pickup."""
    orderId: str
    customerId: str


class PickupOpened(TypedDict):
    """Published Collection fact used to request the customer notification."""
    pickupId: str
    orderId: str
    customerId: str
    collectionCode: str


class OrderCollected(TypedDict):
    """Published Collection fact used to request a loyalty credit."""
    orderId: str
    customerId: str
