"""Collection owns its private command identity, validation and stored wire format."""

from operations.adaptors.generated.message_names import CollectionPrivateCommand, CollectionSubscription

from operations.adaptors.messaging.internal_command_codec import InternalCommandCodec
from operations.adaptors.generated.cafe.internal.collection.internal_commands_pb2 import CommandEnvelope
from operations.contexts.collection.application.commands.open_pickup import OpenPickupCommand
from operations.foundation.domain import identifier, record


def open_command(value: object) -> OpenPickupCommand:
    data = record(value)
    return OpenPickupCommand(orderId=identifier(data["orderId"]), customerId=identifier(data["customerId"]))


def command_codec() -> InternalCommandCodec[OpenPickupCommand]:
    return InternalCommandCodec(
        "collection",
        CollectionSubscription.OPEN_PICKUP,
        CollectionPrivateCommand.OPEN_PICKUP,
        CommandEnvelope,
        open_command,
    )


def command_header(body: bytes) -> tuple[str, str, str]:
    envelope = CommandEnvelope.FromString(body)
    if envelope.consumer != CollectionSubscription.OPEN_PICKUP:
        raise ValueError("Foreign command destination")
    return identifier(envelope.id), envelope.consumer + ".command", identifier(envelope.correlation_id)
