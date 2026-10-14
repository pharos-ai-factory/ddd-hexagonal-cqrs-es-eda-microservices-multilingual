"""Validate schema-declared constraints and serialise published payloads."""

from collections.abc import Callable, Mapping
from dataclasses import dataclass
from datetime import datetime
import re
from google.protobuf.json_format import ParseDict
from google.protobuf.message import Message
from operations.adaptors.generated.cafe.v1.event_options_pb2 import FieldRules, rules
from operations.foundation.domain import identifier


def timestamp(value: str) -> None:
    if not re.fullmatch(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})", value):
        raise ValueError("A timestamp with an offset is required")
    datetime.fromisoformat(value.replace("Z", "+00:00"))


def validate_payload(value: Message, aggregate_id: str) -> None:
    for field in value.DESCRIPTOR.fields:
        rule: FieldRules = field.GetOptions().Extensions[rules]  # type: ignore[index]
        item = getattr(value, field.name)
        if field.is_repeated:
            if (
                rule.HasField("min_items")
                and len(item) < rule.min_items
                or rule.HasField("max_items")
                and len(item) > rule.max_items
            ):
                raise ValueError("Invalid collection size: " + field.name)
            for child in item:
                validate_payload(child, aggregate_id)
        elif field.message_type:
            if not value.HasField(field.name):
                raise ValueError("Missing message: " + field.name)
            validate_payload(item, aggregate_id)
        elif field.type == field.TYPE_STRING:
            if rule.uuid:
                identifier(item)
            if rule.source_identity and item != aggregate_id:
                raise ValueError("Payload does not identify its source aggregate")
            if rule.non_empty and not item.strip() or rule.pattern and not re.fullmatch(rule.pattern, item):
                raise ValueError("Invalid string: " + field.name)
            if rule.timestamp:
                timestamp(item)
        elif field.type != field.TYPE_BOOL:
            if (
                abs(item) > 2**53 - 1
                or rule.HasField("minimum")
                and item < rule.minimum
                or rule.HasField("maximum")
                and item > rule.maximum
            ):
                raise ValueError("Invalid portable integer: " + field.name)


def parse_payload(value: Message, body: bytes, aggregate_id: str) -> None:
    value.ParseFromString(body)
    validate_payload(value, aggregate_id)


@dataclass(frozen=True)
class PayloadCodec:
    """Generated registration binds one published identity to its typed conversion."""

    owner: str
    kind: str
    version: int
    message: type[Message]
    decode: Callable[[bytes, str], Mapping[str, object]]

    def encode(self, payload: Mapping[str, object], aggregate_id: str) -> bytes:
        value = self.message()
        ParseDict(dict(payload), value)
        validate_payload(value, aggregate_id)
        return value.SerializeToString(deterministic=True)
