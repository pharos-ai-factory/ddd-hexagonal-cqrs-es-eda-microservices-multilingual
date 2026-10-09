"""Fixed-value command wire and event-translation evidence, independent of infrastructure."""
from dataclasses import replace
from pathlib import Path
from collections.abc import Mapping
from operations.adaptors.internal_commands import InternalCommandCodec
from operations.contexts.collection.application import OpenPickupCommand
import pytest
from operations.apps.composition.preparation import command_codec
from operations.apps.composition.collection import command_codec as pickup_codec
from operations.contexts.preparation.application import AcceptOrderCommand
from operations.foundation.application import Metadata
from operations.foundation.identity import derived_id


ID = "00000000-0000-4000-8000-000000000001"
SOURCE = "00000000-0000-4000-8000-000000000002"


def test_command_roundtrip_preserves_receipt_material_and_typed_intent() -> None:
    codec = command_codec()
    command = AcceptOrderCommand(orderId=SOURCE, customerId=ID, instructions="2 × Coffee")
    metadata = Metadata(id=derived_id(codec.consumer, SOURCE), target=ID, name=codec.consumer,
        correlation=ID, causation=SOURCE, consumer=codec.consumer, source_id=SOURCE, source_hash="a"*64,
        input={"original": "receipt material", "zero": 0})
    body = codec.encode(metadata, command)
    assert codec.encode(metadata, command) == body
    assert codec.decode(body) == (metadata, command)
    with pytest.raises(ValueError):
        codec.encode(replace(metadata, consumer="collection.foreign"), command)
    with pytest.raises(ValueError):
        codec.encode(replace(metadata, id=ID), command)
    with pytest.raises(ValueError):
        pickup_codec().decode(body)


def test_private_commands_retain_historical_wire_fixtures() -> None:
    def check[C: Mapping[str, object]](codec: InternalCommandCodec[C], command: C) -> None:
        metadata = Metadata(id=derived_id(codec.consumer, SOURCE), target=ID, name=codec.consumer,
            correlation=ID, causation=SOURCE, consumer=codec.consumer, source_id=SOURCE, source_hash="a"*64,
            input={"original": "receipt"})
        path = Path(__file__).parents[1]/"src/operations/contexts"/codec.owner/"adaptors/messaging/fixtures"/(codec.consumer+".command.hex")
        saved = bytes.fromhex(path.read_text().strip())
        assert codec.encode(metadata, command) == saved
        assert codec.decode(saved) == (metadata, command)
    check(command_codec(), AcceptOrderCommand(orderId=SOURCE, customerId=ID, instructions="2 × Coffee"))
    check(pickup_codec(), OpenPickupCommand(orderId=SOURCE, customerId=ID))
