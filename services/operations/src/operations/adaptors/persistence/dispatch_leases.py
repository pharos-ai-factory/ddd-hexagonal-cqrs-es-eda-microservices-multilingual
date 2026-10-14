"""Claim and complete fenced PostgreSQL leases for events, commands and realtime publications."""

from typing import Literal, NotRequired, TypedDict

from operations.adaptors.persistence.postgres_database import PostgresContextDatabase
from operations.foundation.domain import identifier, integer, text
from operations.foundation.identity import new_id


class OutboxDispatch(TypedDict):
    """Carries immutable publication bytes and the fenced lease used to complete one dispatch."""

    id: str
    token: str
    generation: int
    body: bytes
    channel: NotRequired[str]


def binary(value: object) -> bytes:
    if not isinstance(value, (bytes, bytearray, memoryview)):
        raise ValueError("Publication bytes are required")
    return bytes(value)


def claim(
    database: PostgresContextDatabase, realtime: bool | Literal["commands"] = False
) -> OutboxDispatch | None:
    dispatch = (
        "internal_command_dispatches"
        if realtime == "commands"
        else "realtime_dispatches"
        if realtime
        else "dispatches"
    )
    source = (
        "internal_commands"
        if realtime == "commands"
        else "realtime_publications"
        if realtime
        else "outbox_events"
    )
    token = new_id()
    with database.pool.connection() as connection:
        row = connection.execute(
            f"""WITH candidate AS (
            SELECT event_id FROM cafe.{dispatch} WHERE completed_at IS NULL
            AND available_at<=clock_timestamp() AND (lease_until IS NULL OR lease_until<clock_timestamp())
            ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
        ), claimed AS (
            UPDATE cafe.{dispatch} d SET lease_token=%s,lease_until=clock_timestamp()+interval '30 seconds',
            generation=generation+1 FROM candidate c WHERE d.event_id=c.event_id
            RETURNING d.event_id,d.generation
        ) SELECT o.*,c.generation FROM claimed c JOIN cafe.{source} o ON o.id=c.event_id""",
            (token,),
        ).fetchone()
    if not row:
        return None
    result = OutboxDispatch(
        id=identifier(str(row["id"])),
        token=token,
        generation=integer(row["generation"]),
        body=binary(row["body"]),
    )
    if realtime is True:
        result["channel"] = text(row["channel"])
    return result


def finish(
    database: PostgresContextDatabase,
    row: OutboxDispatch,
    realtime: bool | Literal["commands"] = False,
    error: str | None = None,
) -> None:
    table = (
        "internal_command_dispatches"
        if realtime == "commands"
        else "realtime_dispatches"
        if realtime
        else "dispatches"
    )
    with database.pool.connection() as connection:
        connection.execute(
            f"""UPDATE cafe.{table} SET lease_token=NULL,lease_until=NULL,
            available_at=clock_timestamp()+interval '1 second',
            completed_at=CASE WHEN %s::text IS NULL THEN clock_timestamp() ELSE NULL END,last_error=%s
            WHERE event_id=%s AND lease_token=%s AND generation=%s AND lease_until>clock_timestamp()""",
            (error, error, row["id"], row["token"], row["generation"]),
        )
