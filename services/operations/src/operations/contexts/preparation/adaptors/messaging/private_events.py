"""Private delivery codecs and plain domain-event restoration owned by preparation."""

from operations.adaptors.generated.message_names import PreparationDomainEvent

from operations.adaptors.messaging.private_event_codec import PrivateEventCodec
from operations.adaptors.messaging.incoming_event import IncomingEvent
from operations.adaptors.messaging.event_payload_codec import parse_payload
from operations.adaptors.generated.cafe.internal.preparation import private_messages_pb2 as wire
from operations.contexts.preparation.domain.preparation_ticket import DrinksReadyDomainEvent
from operations.foundation.domain import identifier


def decode_drinks_ready(body: bytes, source: str) -> DrinksReadyDomainEvent:
    value = wire.DrinksReadyDomainEvent()
    parse_payload(value, body, source)
    identifier(value.order_id)
    identifier(value.customer_id)
    return DrinksReadyDomainEvent(value.order_id, value.customer_id)


def validate_drinks_ready(body: bytes, source: str) -> dict[str, object]:
    event = decode_drinks_ready(body, source)
    return {"order_id": event.order_id, "customer_id": event.customer_id}


DRINKS_READY_DOMAIN = IncomingEvent[DrinksReadyDomainEvent](
    PreparationDomainEvent.DRINKS_READY, lambda event: decode_drinks_ready(event.payload, event.aggregate_id)
)

PRIVATE_CODECS = {
    PreparationDomainEvent.DRINKS_READY: PrivateEventCodec(
        "preparation",
        PreparationDomainEvent.DRINKS_READY,
        "ticket",
        wire.PrivateEvent,
        "drinks_ready",
        validate_drinks_ready,
    ),
}
