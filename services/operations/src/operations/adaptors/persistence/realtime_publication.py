"""Persist explicit browser publications using the caller's transaction."""

from collections.abc import Mapping, Sequence
import re
from psycopg import Connection
from operations.adaptors.persistence.postgres_database import Row
from operations.adaptors.realtime import message_codec
from operations.foundation.application import Metadata, Outcome, RealtimeMessage


def realtime_channel(owner: str, event: RealtimeMessage) -> str:
    """Resolve an application target after validating its payload authority."""
    if event.target.kind == "context":
        if event.target.id != owner or event.snapshot is None or event.rejection is not None:
            raise ValueError("Invalid context realtime target")
        return "cafe:" + owner
    if event.target.kind == "principal":
        if (
            event.snapshot is not None
            or event.rejection is None
            or not re.fullmatch(
                r"[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}",
                event.target.id,
            )
        ):
            raise ValueError("Invalid principal realtime target")
        return "principal:" + event.target.id
    raise ValueError("Invalid realtime delivery target")


def persist_realtime(
    connection: Connection[Row],
    owner: str,
    kind: str,
    m: Metadata,
    outcome: Outcome,
    pending: Mapping[str, object] | None,
    events: Sequence[RealtimeMessage],
) -> None:
    """Bind each explicit message to the snapshot or rejection being committed."""
    for event in events:
        message_kind = "snapshot"
        if event.snapshot is not None:
            if "rejection" in outcome:
                continue
            if pending is None or dict(event.snapshot) != dict(pending):
                raise ValueError("Realtime snapshot differs from committed state")
            identity, body = message_codec.encode(
                owner, kind, m.target, outcome["version"], event.target, event.snapshot
            )
        else:
            if event.rejection is None or outcome.get("rejection") != event.rejection:
                raise ValueError("Realtime rejection differs from committed outcome")
            if not m.actor_id or event.target.id != m.actor_id:
                raise ValueError("Realtime rejection target differs from authenticated actor")
            message_kind = "operation_rejected"
            identity, body = message_codec.encode_rejection(
                owner,
                kind,
                m.target,
                outcome["version"],
                event.target,
                m.id,
                m.correlation,
                event.rejection["code"],
                event.rejection["message"],
            )
        append_realtime(
            connection,
            realtime_channel(owner, event),
            kind,
            m.target,
            outcome["version"],
            identity,
            body,
            message_kind,
        )


def append_realtime(
    connection: Connection[Row],
    channel: str,
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
        (identity, channel, kind, target, version, body, message_kind),
    )
    connection.execute("INSERT INTO cafe.realtime_dispatches(event_id) VALUES(%s)", (identity,))
