"""Typed incoming published fact accepted by Preparation."""
from typing import Final
from operations.adaptors import codec
from operations.contracts.events import OrderPlaced
from operations.adaptors.incoming_event import IncomingEvent

ORDER_PLACED: Final[IncomingEvent[OrderPlaced]] = IncomingEvent('ordering.order-placed', codec.order_placed)
