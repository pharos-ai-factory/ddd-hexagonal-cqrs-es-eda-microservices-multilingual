"""Persist explicit browser publications using the caller's transaction."""

from collections.abc import Mapping, Sequence
from psycopg import Connection
from operations.adaptors.persistence.postgres_database import Row
from operations.adaptors.realtime import projection_codec
from operations.foundation.application import Metadata, Outcome, RealtimeEvent


def persist_realtime(
    connection: Connection[Row],
    owner: str,
    kind: str,
    m: Metadata,
    outcome: Outcome,
    pending: Mapping[str, object] | None,
    events: Sequence[RealtimeEvent],
) -> None:
    """Bind each explicit message to the snapshot or rejection being committed."""
    for event in events:
        message_kind = "snapshot"
        if event.snapshot is not None:
            if "rejection" in outcome:
                continue
            if pending is None or dict(event.snapshot) != dict(pending):
                raise ValueError("Realtime snapshot differs from committed state")
            identity, body = projection_codec.encode(
                owner, kind, m.target, outcome["version"], event.snapshot
            )
        else:
            if event.rejection is None or outcome.get("rejection") != event.rejection:
                raise ValueError("Realtime rejection differs from committed outcome")
            message_kind = "operation_rejected"
            identity, body = projection_codec.encode_rejection(
                owner,
                kind,
                m.target,
                outcome["version"],
                m.id,
                m.correlation,
                event.rejection["code"],
                event.rejection["message"],
            )
        append_realtime(connection, owner, kind, m.target, outcome["version"], identity, body, message_kind)


def append_realtime(
    connection: Connection[Row],
    owner: str,
    kind: str,
    target: str,
    version: int,
    identity: str,
    body: bytes,
    message_kind: str = "snapshot",
) -> None:
    """Stage exact bytes and a dispatch together on the active owner transaction."""
    connection.execute(
        """INSERT INTO cafe.realtime_publications
        (id,channel,aggregate_kind,aggregate_id,revision,body,message_kind) VALUES(%s,%s,%s,%s,%s,%s,%s)""",
        (identity, "cafe:" + owner, kind, target, version, body, message_kind),
    )
    connection.execute("INSERT INTO cafe.realtime_dispatches(event_id) VALUES(%s)", (identity,))
