"""Preparation owns its private command identity, validation and stored wire format."""

from operations.adaptors.generated.message_names import PreparationPrivateCommand, PreparationSubscription

from operations.adaptors.messaging.internal_command_codec import InternalCommandCodec
from operations.adaptors.generated.cafe.internal.preparation.internal_commands_pb2 import CommandEnvelope
from operations.contexts.preparation.application.commands.accept_order import AcceptOrderCommand
from operations.foundation.domain import identifier, record, text


def accept_command(value: object) -> AcceptOrderCommand:
    data = record(value)
    return AcceptOrderCommand(
        orderId=identifier(data["orderId"]),
        customerId=identifier(data["customerId"]),
        instructions=text(data["instructions"]),
    )


def command_codec() -> InternalCommandCodec[AcceptOrderCommand]:
    return InternalCommandCodec(
        "preparation",
        PreparationSubscription.ACCEPT_ORDER,
        PreparationPrivateCommand.ACCEPT_ORDER,
        CommandEnvelope,
        accept_command,
    )


def command_header(body: bytes) -> tuple[str, str, str]:
    envelope = CommandEnvelope.FromString(body)
    if envelope.consumer != PreparationSubscription.ACCEPT_ORDER:
        raise ValueError("Foreign command destination")
    return identifier(envelope.id), envelope.consumer + ".command", identifier(envelope.correlation_id)
