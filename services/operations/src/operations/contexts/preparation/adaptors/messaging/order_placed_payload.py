"""Check immutable accepted-order evidence before Preparation reacts."""

from operations.adaptors.generated.message_names import OrderingIntegrationEvent

from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.adaptors.generated.event_ordering import decode_OrderPlaced
from operations.contracts.events.generated.ordering import OrderPlacedIntegrationEvent


def order_placed(event: Event) -> OrderPlacedIntegrationEvent:
    if event.name != OrderingIntegrationEvent.ORDER_PLACED:
        raise ValueError("Expected OrderPlacedIntegrationEvent")
    payload = decode_OrderPlaced(event.payload, event.aggregate_id)
    lines = payload["lines"]
    if (
        len({line["id"] for line in lines}) != len(lines)
        or not 1 <= sum(line["quantity"] for line in lines) <= 5
    ):
        raise ValueError("Invalid accepted order evidence")
    return payload
