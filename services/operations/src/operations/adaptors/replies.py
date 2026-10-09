"""Transaction-scoped command reply intent and fenced confirmed publication."""
from collections.abc import Callable
from contextvars import ContextVar
from dataclasses import dataclass
from threading import Event
from typing import TYPE_CHECKING
import pika
from operations.foundation.application import Outcome
from operations.foundation.identity import new_id
if TYPE_CHECKING:
    from psycopg import Connection
    from operations.adaptors.postgres import PostgresContextDatabase, Row


@dataclass
class ReplyIntent:
    """Tracks the exact response to commit alongside the receiving command transaction."""
    identity: str
    encode: Callable[[Outcome], bytes]
    persisted: bool = False


current: ContextVar[ReplyIntent | None] = ContextVar("command_reply", default=None)


def append(connection: "Connection[Row]", outcome: Outcome) -> Outcome:
    intent = current.get()
    if intent is not None:
        body = intent.encode(outcome)
        if not body:
            raise ValueError("Empty command reply")
        connection.execute("INSERT INTO cafe.command_replies(id,body) VALUES(%s,%s) ON CONFLICT DO NOTHING",
                           (intent.identity, body))
        row = connection.execute("SELECT body FROM cafe.command_replies WHERE id=%s", (intent.identity,)).fetchone()
        if row is None or bytes(row["body"]) != body:  # type: ignore[call-overload]
            raise ValueError("Request identity reused with different reply bytes")
        connection.execute("INSERT INTO cafe.command_reply_dispatches(event_id) VALUES(%s) ON CONFLICT DO NOTHING",
                           (intent.identity,))
        intent.persisted = True
    return outcome


def claim_reply(database: "PostgresContextDatabase", token: str) -> "Row | None":
    with database.pool.connection() as connection:
        return connection.execute("""WITH candidate AS (
            SELECT event_id FROM cafe.command_reply_dispatches WHERE completed_at IS NULL
            AND available_at<=clock_timestamp() AND (lease_until IS NULL OR lease_until<clock_timestamp())
            ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
        ), claimed AS (
            UPDATE cafe.command_reply_dispatches d SET lease_token=%s,
            lease_until=clock_timestamp()+interval '8 seconds',generation=generation+1,attempts=attempts+1
            FROM candidate c WHERE d.event_id=c.event_id RETURNING d.event_id,d.generation
        ) SELECT o.*,c.generation,o.expires_at<=clock_timestamp() AS expired
            FROM claimed c JOIN cafe.command_replies o ON o.id=c.event_id""", (token,)).fetchone()


def relay(database: "PostgresContextDatabase", url: str, stop: Event) -> None:
    from operations.adaptors.delivery import broker_connection, binary
    from operations.adaptors.diagnostics import failure
    while not stop.is_set():
        try:
            with broker_connection(url) as broker:
                channel = broker.channel()
                channel.confirm_delivery()
                while not stop.is_set():
                    token = new_id()
                    row = claim_reply(database, token)
                    if row is None:
                        broker.process_data_events(time_limit=0.1)  # type: ignore[arg-type]
                        continue
                    reason = None
                    try:
                        if row["expired"]:
                            reason = "expired"
                        else:
                            identity = str(row["id"])
                            channel.basic_publish("cafe.replies", "reply."+database.owner, binary(row["body"]),
                                properties=pika.BasicProperties(content_type="application/x-protobuf", delivery_mode=2,
                                    type="reply", app_id=database.owner, message_id=identity, correlation_id=identity), mandatory=True)
                    except Exception:
                        reason = "publication_failed"
                        raise
                    finally:
                        with database.pool.connection() as connection:
                            connection.execute("""UPDATE cafe.command_reply_dispatches SET lease_token=NULL,lease_until=NULL,
                                available_at=clock_timestamp()+interval '1 second',last_error=%s,
                                completed_at=CASE WHEN %s::text IS NULL OR %s='expired' THEN clock_timestamp() ELSE NULL END
                                WHERE event_id=%s AND lease_token=%s AND generation=%s AND lease_until>clock_timestamp()""",
                                (reason, reason, reason, row["id"], token, row["generation"]))
        except Exception as error:
            failure(database.owner, "replies.reconnect", error)
            stop.wait(1)
