"""Typed incoming published fact accepted by Collection."""

from typing import Final
from operations.adaptors.generated.event_preparation import decode_DrinksReady
from operations.contracts.events.generated.preparation import DrinksReady
from operations.adaptors.incoming_event import IncomingEvent

DRINKS_READY: Final[IncomingEvent[DrinksReady]] = IncomingEvent(
    "preparation.drinks-ready", lambda event: decode_DrinksReady(event.payload, event.aggregate_id)
)
