"""Typed incoming published fact accepted by Preparation."""

from typing import Final
from operations.contexts.preparation.adaptors.messaging.order_placed_payload import order_placed
from operations.contracts.events.ordering import OrderPlaced
from operations.adaptors.incoming_event import IncomingEvent

ORDER_PLACED: Final[IncomingEvent[OrderPlaced]] = IncomingEvent("ordering.order-placed", order_placed)
