"""Validate generated request envelopes and construct typed outcome replies."""

from google.protobuf.message import Message
from operations.adaptors.messaging.generated.response_payloads import to_outcome

from operations.adaptors.generated.cafe.requests.v1.validation_pb2 import required_input
from operations.adaptors.generated.request_registry import REPLY_TYPES, REQUEST_TYPES, Reply, Request
from operations.foundation.application import Outcome
from operations.foundation.domain import identifier, integer


def new_reply(owner: str) -> Reply:
    factory = REPLY_TYPES.get(owner)
    if factory is None:
        raise ValueError("Unknown request owner")
    return factory()


def parse_request(owner: str, body: bytes) -> Request:
    factory = REQUEST_TYPES.get(owner)
    if factory is None:
        raise ValueError("Unknown request owner")
    return factory.FromString(body)


def payload(message: Message) -> tuple[str, Message]:
    for field, value in message.ListFields():
        if field.containing_oneof and field.containing_oneof.name == "payload" and isinstance(value, Message):
            return field.json_name, value
    raise ValueError("A typed payload is required")


def validate(request: Request, owner: str) -> tuple[str, Message]:
    if request.contract_version != 1 or request.context != owner:
        raise ValueError("Invalid request envelope")
    identifier(request.request_id)
    kind, outer = payload(request)
    name, body = payload(outer)
    if "/contexts/" + owner + "/" not in body.DESCRIPTOR.file.name:
        raise ValueError("Foreign context request")
    if kind == "command":
        metadata = request.command.metadata
        identifier(metadata.command_id)
        identifier(metadata.aggregate_id)
        identifier(metadata.correlation_id)
        if not metadata.HasField("expected_version"):
            raise ValueError("Expected version is required")
        required_inputs(body)
    else:
        if "id" in body.DESCRIPTOR.fields_by_name:
            identifier(getattr(body, "id"))
        if "page" in body.DESCRIPTOR.fields_by_name and body.HasField("page"):
            page = getattr(body, "page")
            limit = integer(page.limit)
            if not 1 <= limit <= 100:
                raise ValueError("Invalid page limit")
            if page.HasField("after"):
                identifier(page.after)
    return name, body


def outcome_reply(owner: str, outcome: Outcome) -> Reply:
    reply = new_reply(owner)
    reply.outcome.CopyFrom(to_outcome(outcome))
    return reply


def required_inputs(body: Message) -> None:
    for field in body.DESCRIPTOR.fields:
        required = any(
            option.full_name == required_input.full_name and value is True
            for option, value in field.GetOptions().ListFields()
        )
        if required and not body.HasField(field.name):
            raise ValueError("A required command field is absent")
