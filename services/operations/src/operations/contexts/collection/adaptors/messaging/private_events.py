"""Private delivery codecs and plain domain-event restoration owned by collection."""

from operations.adaptors.private_event import PrivateEventCodec
from operations.adaptors.incoming_event import IncomingEvent
from operations.adaptors.event_payload import parse_payload
from operations.adaptors.generated.cafe.internal.collection import private_messages_pb2 as wire
from operations.contexts.collection.domain.pickup import PickupOpenedDomainEvent, OrderCollectedDomainEvent
from operations.foundation.domain import identifier
from operations.contexts.collection.domain.pickup import CollectionCode


def decode_pickup_opened(body: bytes, source: str) -> PickupOpenedDomainEvent:
    value = wire.PickupOpenedDomainEvent()
    parse_payload(value, body, source)
    identifier(value.pickup_id)
    identifier(value.order_id)
    identifier(value.customer_id)
    if value.pickup_id != source:
        raise ValueError("Pickup source identity mismatch")
    CollectionCode(value.collection_code)
    return PickupOpenedDomainEvent(value.pickup_id, value.order_id, value.customer_id, value.collection_code)


def validate_pickup_opened(body: bytes, source: str) -> dict[str, object]:
    event = decode_pickup_opened(body, source)
    return {
        "pickup_id": event.pickup_id,
        "order_id": event.order_id,
        "customer_id": event.customer_id,
        "collection_code": event.code,
    }


PICKUP_OPENED_DOMAIN = IncomingEvent[PickupOpenedDomainEvent](
    "collection.pickup-opened.domain", lambda event: decode_pickup_opened(event.payload, event.aggregate_id)
)


def decode_order_collected(body: bytes, source: str) -> OrderCollectedDomainEvent:
    value = wire.OrderCollectedDomainEvent()
    parse_payload(value, body, source)
    identifier(value.order_id)
    identifier(value.customer_id)
    return OrderCollectedDomainEvent(value.order_id, value.customer_id)


def validate_order_collected(body: bytes, source: str) -> dict[str, object]:
    event = decode_order_collected(body, source)
    return {"order_id": event.order_id, "customer_id": event.customer_id}


ORDER_COLLECTED_DOMAIN = IncomingEvent[OrderCollectedDomainEvent](
    "collection.order-collected.domain",
    lambda event: decode_order_collected(event.payload, event.aggregate_id),
)

PRIVATE_CODECS = {
    "collection.pickup-opened.domain": PrivateEventCodec(
        "collection",
        "collection.pickup-opened.domain",
        "pickup",
        wire.PrivateEvent,
        "pickup_opened",
        validate_pickup_opened,
    ),
    "collection.order-collected.domain": PrivateEventCodec(
        "collection",
        "collection.order-collected.domain",
        "pickup",
        wire.PrivateEvent,
        "order_collected",
        validate_order_collected,
    ),
}
