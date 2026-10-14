"""Versioned wire mapping. Generated messages never enter domain packages."""

from datetime import datetime, timezone
from collections.abc import Mapping
from google.protobuf.json_format import ParseDict
from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.adaptors.generated.cafe.realtime.v1.realtime_pb2 import Publication as Realtime
from operations.foundation.application import Metadata, Publication
from operations.foundation.domain import identifier
from operations.foundation.identity import new_id

from operations.adaptors.generated.event_registry import PAYLOAD_CODECS
from operations.adaptors.event_payload import timestamp


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
    validate(event)
    return event, dict(PAYLOAD_CODECS[event.name].decode(event.payload, event.aggregate_id))


def encode(
    owner: str, kind: str, identity: str, version: int, metadata: Metadata, publication: Publication
) -> tuple[Event, bytes]:
    definition = PAYLOAD_CODECS[publication.name]
    event = Event(
        id=new_id(),
        name=publication.name,
        context=owner,
        visibility="integration",
        contract_version=definition.version,
        aggregate_kind=kind,
        aggregate_id=identity,
        aggregate_version=version,
        correlation_id=metadata.correlation,
        causation_id=metadata.id,
        occurred_at=datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
    )
    event.payload = definition.encode(publication.payload, identity)
    validate(event)
    body = event.SerializeToString(deterministic=True)
    if len(body) > 256 * 1024:
        raise ValueError("Event exceeds its wire limit")
    return event, body


def realtime(
    owner: str, kind: str, identity: str, version: int, state: Mapping[str, object]
) -> tuple[str, bytes]:
    event = Realtime(
        event_id=new_id(),
        contract_version=1,
        context=owner,
        aggregate_kind=kind,
        aggregate_id=identity,
        revision=version,
    )
    ParseDict(dict(state), getattr(event, kind))
    return event.event_id, event.SerializeToString(deterministic=True)
