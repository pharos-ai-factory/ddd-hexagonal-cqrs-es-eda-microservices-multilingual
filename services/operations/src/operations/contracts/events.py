"""The published event shapes consumed or emitted by Operations."""
from typing import TypedDict


class OrderLine(TypedDict):
    id: str
    offerCode: str
    name: str
    quantity: int
    minor: int


class OrderPlaced(TypedDict):
    orderId: str
    customerId: str
    editionId: str
    currency: str
    lines: list[OrderLine]


class DrinksReady(TypedDict):
    orderId: str
    customerId: str


class PickupOpened(TypedDict):
    pickupId: str
    orderId: str
    customerId: str
    collectionCode: str


class OrderCollected(TypedDict):
    orderId: str
    customerId: str
