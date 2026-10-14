"""Typed incoming published fact accepted by Preparation."""

from operations.adaptors.generated.message_names import OrderingIntegrationEvent

from typing import Final
from operations.contexts.preparation.adaptors.messaging.order_placed_payload import order_placed
from operations.contracts.events.generated.ordering import OrderPlacedIntegrationEvent
from operations.adaptors.messaging.incoming_event import IncomingEvent

ORDER_PLACED: Final[IncomingEvent[OrderPlacedIntegrationEvent]] = IncomingEvent(
    OrderingIntegrationEvent.ORDER_PLACED, order_placed
)
