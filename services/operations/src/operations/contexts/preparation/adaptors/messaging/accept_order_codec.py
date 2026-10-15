"""Preparation owns its private command identity, validation and stored wire format."""

from .generated.request_payloads import parse_accept_order_command


from operations.adaptors.generated.message_names import PreparationPrivateCommand, PreparationSubscription

from operations.adaptors.messaging.internal_command_codec import InternalCommandCodec
from operations.adaptors.generated.cafe.internal.preparation.internal_commands_pb2 import CommandEnvelope
from operations.contracts.requests.generated.preparation import AcceptOrderCommand
from operations.foundation.domain import identifier, text


def accept_command(value: object) -> AcceptOrderCommand:
    command = parse_accept_order_command(value)
    identifier(command["orderId"])
    identifier(command["customerId"])
    text(command["instructions"])
    return command


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
