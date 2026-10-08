"""Native transport translation checks; delivery claims use real RabbitMQ."""
import pytest
from operations.contexts.collection.adaptors.messaging import requests as collection_wire
from operations.adaptors.requests import Registry, validate
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_requests_pb2 import Request
from operations.foundation.application import Metadata, Outcome
from operations.contexts.collection.application import CollectOrderCommand

ID = "11111111-1111-4111-8111-111111111111"


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
    registry = Registry("collection")
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
    from operations.adaptors.requests import required_inputs
    source = descriptor_pb2.FileDescriptorProto(name="future_python_input.proto", package="test", syntax="proto3")
    message = source.message_type.add(name="Input")
    for index, name in enumerate(("price", "note")):
        message.oneof_decl.add(name="_"+name)
        field = message.field.add(name=name, number=index+1, type=5 if index == 0 else 9,
                                  label=1, oneof_index=index, proto3_optional=True)
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
