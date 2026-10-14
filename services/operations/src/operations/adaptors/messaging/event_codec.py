"""Encode and validate delivered domain and integration event envelopes."""

from datetime import datetime, timezone

from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.adaptors.generated.event_registry import PAYLOAD_CODECS
from operations.adaptors.generated.private_event_registry import PRIVATE_CODECS
from operations.adaptors.messaging.event_payload_codec import timestamp
from operations.foundation.application import Metadata, Publication
from operations.foundation.domain import identifier
from operations.foundation.identity import new_id


def validate(event: Event) -> None:
    definition = PAYLOAD_CODECS.get(event.name)
    if (
        definition is None
        or event.context != definition.owner
        or event.visibility != "integration"
        or event.aggregate_kind != definition.kind
        or event.contract_version != definition.version
    ):
        raise ValueError("Unknown event contract or owner")
    for value in (event.id, event.aggregate_id, event.correlation_id, event.causation_id):
        identifier(value)
    if event.aggregate_version < 1 or event.aggregate_version > 2**53 - 1 or not event.payload:
        raise ValueError("Missing payload or invalid aggregate version")
    timestamp(event.occurred_at)


def decode(body: bytes) -> tuple[Event, dict[str, object]]:
    if len(body) > 256 * 1024:
        raise ValueError("Event exceeds its wire limit")
    event = Event.FromString(body)
    if event.visibility == "domain":
        private = PRIVATE_CODECS.get(event.name)
        if private is None:
            raise ValueError("Unknown private event")
        return private.decode(body)
    validate(event)
    return event, dict(PAYLOAD_CODECS[event.name].decode(event.payload, event.aggregate_id))


def encode(
    owner: str,
    kind: str,
    identity: str,
    version: int,
    metadata: Metadata,
    publication: Publication,
    *,
    occurred_at: str | None = None,
) -> tuple[Event, bytes]:
    private = PRIVATE_CODECS.get(publication.name)
    definition = PAYLOAD_CODECS.get(publication.name)
    if private is None and definition is None:
        raise ValueError("Unknown publication")
    event = Event(
        id=new_id(),
        name=publication.name,
        context=owner,
        visibility="domain" if private else "integration",
        contract_version=definition.version if definition else 1,
        aggregate_kind=kind,
        aggregate_id=identity,
        aggregate_version=version,
        correlation_id=metadata.correlation,
        causation_id=metadata.id,
        occurred_at=occurred_at or datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
    )
    if private:
        body = private.encode(event, publication.payload)
        if len(body) > 256 * 1024:
            raise ValueError("Event exceeds its wire limit")
        return event, body
    assert definition is not None
    event.payload = definition.encode(publication.payload, identity)
    validate(event)
    body = event.SerializeToString(deterministic=True)
    if len(body) > 256 * 1024:
        raise ValueError("Event exceeds its wire limit")
    return event, body
