"""Native transport translation checks; delivery claims use real RabbitMQ."""

import pytest
from operations.contexts.collection.adaptors.messaging import pickup_requests as collection_wire
from operations.adaptors.messaging.rabbitmq_request_registry import RabbitMQRequestRegistry
from operations.adaptors.messaging.request_envelope import validate
from operations.adaptors.messaging.request_envelope import new_reply, parse_request
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_requests_pb2 import Request
from operations.foundation.application import Metadata, Outcome
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommand

ID = "11111111-1111-4111-8111-111111111111"


@pytest.mark.parametrize("owner", ["preparation", "collection"])
def test_generated_owner_envelopes_preserve_request_identity(owner: str) -> None:
    reply = new_reply(owner)
    assert reply.DESCRIPTOR.full_name == f"cafe.{owner}.requests.v1.Reply"
    from operations.adaptors.generated.request_registry import REQUEST_TYPES

    request = REQUEST_TYPES[owner](contract_version=1, request_id=ID, context=owner)
    decoded = parse_request(owner, request.SerializeToString(deterministic=True))
    assert decoded == request
    assert decoded.DESCRIPTOR.full_name == f"cafe.{owner}.requests.v1.Request"


def test_unknown_request_owner_is_rejected() -> None:
    with pytest.raises(ValueError, match="Unknown request owner"):
        parse_request("unknown", b"")
    with pytest.raises(ValueError, match="Unknown request owner"):
        new_reply("unknown")


def command() -> Request:
    request = Request(contract_version=1, request_id=ID, context="collection")
    request.command.metadata.command_id = ID
    request.command.metadata.aggregate_id = ID
    request.command.metadata.correlation_id = ID
    request.command.metadata.expected_version = 0
    request.command.collect_order.code = "1234"
    return request


def test_wire_command_maps_to_plain_input_and_preserves_zero_version() -> None:
    captured: list[Metadata] = []

    def execute(metadata: Metadata, value: CollectOrderCommand) -> Outcome:
        assert value == {"code": "1234"}
        captured.append(metadata)
        return Outcome(aggregateId=ID, version=1, status="collected")

    registry = RabbitMQRequestRegistry("collection")
    registry.command("collectOrder", collection_wire.collect, execute)
    reply = registry.handle(Request.FromString(command().SerializeToString()))
    assert captured[0].expected == 0
    assert captured[0].name == "collection.CollectOrder"
    assert captured[0].input == {"code": "1234"}
    assert reply.outcome.version == 1


def test_missing_fields_and_foreign_context_fail_before_application() -> None:
    request = command()
    with pytest.raises(ValueError):
        validate(request, "preparation")
    request.command.metadata.ClearField("expected_version")
    with pytest.raises(ValueError):
        validate(request, "collection")
    request = command()
    request.command.collect_order.ClearField("code")
    with pytest.raises(ValueError):
        validate(request, "collection")


def test_future_optional_input_is_not_required_and_explicit_zero_is_present() -> None:
    from google.protobuf import descriptor_pb2, descriptor_pool, message_factory
    from operations.adaptors.generated.cafe.requests.v1.validation_pb2 import required_input
    from operations.adaptors.messaging.request_envelope import required_inputs

    source = descriptor_pb2.FileDescriptorProto(
        name="future_python_input.proto", package="test", syntax="proto3"
    )
    message = source.message_type.add(name="Input")
    for index, name in enumerate(("price", "note")):
        message.oneof_decl.add(name="_" + name)
        field = message.field.add(
            name=name,
            number=index + 1,
            type=5 if index == 0 else 9,
            label=1,
            oneof_index=index,
            proto3_optional=True,
        )
        if index == 0:
            field.options.Extensions[required_input] = True  # type: ignore[index]
    descriptor_pool.Default().Add(source)
    descriptor = descriptor_pool.Default().FindMessageTypeByName("test.Input")
    value = message_factory.GetMessageClass(descriptor)()
    with pytest.raises(ValueError):
        required_inputs(value)
    setattr(value, "price", 0)
    required_inputs(value)
    setattr(value, "note", "Optional addition")
    required_inputs(value)
