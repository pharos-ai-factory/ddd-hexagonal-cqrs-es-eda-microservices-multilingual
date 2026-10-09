"""Collection owns its private command identity, validation and stored wire format."""
from enum import StrEnum
from operations.adaptors.internal_commands import InternalCommandCodec
from operations.adaptors.generated.cafe.internal.collection.internal_commands_pb2 import CommandEnvelope
from operations.contexts.collection.application.commands.open_pickup import OpenPickupCommand
from operations.foundation.domain import identifier, record


class CollectionSubscription(StrEnum):
    """Stable queue and receipt identity, independent of Python class names."""
    OPEN_PICKUP = "collection.open-pickup"


def open_command(value: object) -> OpenPickupCommand:
    data = record(value)
    return OpenPickupCommand(orderId=identifier(data["orderId"]), customerId=identifier(data["customerId"]))


def command_codec() -> InternalCommandCodec[OpenPickupCommand]:
    return InternalCommandCodec("collection", CollectionSubscription.OPEN_PICKUP, "open_pickup", CommandEnvelope, open_command)


def command_header(body: bytes) -> tuple[str, str, str]:
    envelope = CommandEnvelope.FromString(body)
    if envelope.consumer != CollectionSubscription.OPEN_PICKUP:
        raise ValueError("Foreign command destination")
    return identifier(envelope.id), envelope.consumer+".command", identifier(envelope.correlation_id)
