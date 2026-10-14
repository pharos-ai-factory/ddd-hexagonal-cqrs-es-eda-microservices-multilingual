"""Technical private envelope codec; owner adaptors supply payload validation."""

from collections.abc import Callable, Mapping

from google.protobuf.json_format import ParseDict
from google.protobuf.message import Message

from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.adaptors.messaging.event_payload_codec import timestamp
from operations.foundation.domain import identifier


class PrivateEventCodec:
    """Keep private payload fields and validators inside their owning context."""

    def __init__(
        self,
        owner: str,
        name: str,
        kind: str,
        envelope: type[Message],
        field: str,
        parse: Callable[[bytes, str], Mapping[str, object]],
    ) -> None:
        self.owner, self.name, self.kind = owner, name, kind
        self.envelope, self.field, self.parse = envelope, field, parse

    def validate(self, event: Event) -> None:
        if (
            event.context != self.owner
            or event.name != self.name
            or event.aggregate_kind != self.kind
            or event.visibility != "domain"
            or event.contract_version != 1
        ):
            raise ValueError("Foreign private event contract")
        for value in (event.id, event.aggregate_id, event.correlation_id, event.causation_id):
            identifier(value)
        if not 1 <= event.aggregate_version <= 2**53 - 1:
            raise ValueError("Invalid source revision")
        timestamp(event.occurred_at)

    def decode(self, body: bytes) -> tuple[Event, dict[str, object]]:
        private = self.envelope.FromString(body)
        if private.WhichOneof("payload") != self.field:
            raise ValueError("Wrong private event payload")
        event = Event()
        # Metadata field numbers 1–11 are shared; field 5 is the private format revision.
        for field in event.DESCRIPTOR.fields:
            if field.number <= 11:
                name = "format_revision" if field.number == 5 else field.name
                setattr(event, field.name, getattr(private, name))
        event.payload = getattr(private, self.field).SerializeToString(deterministic=True)
        self.validate(event)
        return event, dict(self.parse(event.payload, event.aggregate_id))

    def encode(self, event: Event, payload: Mapping[str, object]) -> bytes:
        self.validate(event)
        private = self.envelope()
        for field in event.DESCRIPTOR.fields:
            if field.number <= 11:
                name = "format_revision" if field.number == 5 else field.name
                setattr(private, name, getattr(event, field.name))
        value = getattr(private, self.field)
        ParseDict(dict(payload), value)
        self.parse(value.SerializeToString(deterministic=True), event.aggregate_id)
        return private.SerializeToString(deterministic=True)
