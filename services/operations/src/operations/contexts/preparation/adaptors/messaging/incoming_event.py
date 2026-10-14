"""Typed incoming published fact accepted by Preparation."""

from typing import Final
from operations.contexts.preparation.adaptors.messaging.order_placed_payload import order_placed
from operations.contracts.events.generated.ordering import OrderPlacedIntegrationEvent
from operations.adaptors.incoming_event import IncomingEvent

ORDER_PLACED: Final[IncomingEvent[OrderPlacedIntegrationEvent]] = IncomingEvent(
    "ordering.order-placed", order_placed
)
