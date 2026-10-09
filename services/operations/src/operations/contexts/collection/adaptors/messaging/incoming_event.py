"""Typed incoming published fact accepted by Collection."""
from typing import Final
from operations.adaptors import codec
from operations.contracts.events import DrinksReady
from operations.adaptors.incoming_event import IncomingEvent

DRINKS_READY: Final[IncomingEvent[DrinksReady]] = IncomingEvent('preparation.drinks-ready', codec.drinks_ready)
