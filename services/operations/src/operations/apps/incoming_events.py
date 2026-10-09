"""Typed event definitions couple each wire name to its boundary decoder."""
from typing import Final
from operations.adaptors import codec
from operations.contracts.events import OrderPlaced, DrinksReady
from operations.apps.runtime import IncomingEvent

ORDER_PLACED: Final[IncomingEvent[OrderPlaced]] = IncomingEvent('ordering.order-placed', codec.order_placed)
DRINKS_READY: Final[IncomingEvent[DrinksReady]] = IncomingEvent('preparation.drinks-ready', codec.drinks_ready)
