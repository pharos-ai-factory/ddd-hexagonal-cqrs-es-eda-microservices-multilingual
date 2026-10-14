"""Encode authoritative browser snapshots into the separate realtime contract."""

from collections.abc import Mapping

from google.protobuf.json_format import ParseDict

from operations.adaptors.generated.cafe.realtime.v1.realtime_pb2 import Publication as Realtime
from operations.foundation.identity import new_id


def encode(
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
