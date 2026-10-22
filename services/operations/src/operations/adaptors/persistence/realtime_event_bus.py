"""Durable browser publication by explicitly registered event subscribers."""

import re
from psycopg.types.json import Jsonb
from operations.adaptors.persistence.postgres_database import PostgresContextDatabase, check_root_identity
from operations.adaptors.persistence.realtime_publication import append_realtime
from operations.adaptors.realtime.message_codec import encode
from operations.foundation.application import RealtimeMessage, Outcome, context_target
from operations.foundation.domain import identifier


class PostgresRealtimeMessageBus:
    """Commit one subscription's receipt and browser bytes before its delivery is acknowledged."""

    def __init__(self, database: PostgresContextDatabase, consumer: str, kind: str) -> None:
        if not consumer.startswith(database.owner + "."):
            raise ValueError("Foreign realtime subscription")
        self.database, self.consumer, self.kind = database, consumer, kind

    def publish(self, event: RealtimeMessage) -> None:
        m = event.source
        if (
            m is None
            or event.snapshot is None
            or event.rejection is not None
            or not 1 <= event.revision <= 2**53 - 1
        ):
            raise ValueError("Invalid realtime delivery")
        if event.target != context_target(self.database.owner):
            raise ValueError("Invalid realtime delivery target")
        for value in (m.id, m.target, m.source_id, m.correlation):
            identifier(value)
        if m.consumer != self.consumer or not re.fullmatch(r"[a-f0-9]{64}", m.source_hash):
            raise ValueError("Invalid realtime receipt")
        check_root_identity(event.snapshot, m.target)
        target = "realtime:" + self.kind + ":" + m.target
        with self.database.pool.connection() as connection, connection.transaction():
            connection.execute(
                "SELECT pg_advisory_xact_lock(hashtextextended(%s,0))",
                ("consumer:" + self.consumer + ":" + m.source_id,),
            )
            previous = connection.execute(
                "SELECT fingerprint,target FROM cafe.consumer_receipts WHERE consumer=%s AND event_id=%s",
                (self.consumer, m.source_id),
            ).fetchone()
            if previous:
                if previous["fingerprint"] != m.source_hash or previous["target"] != target:
                    raise ValueError("Conflicting realtime delivery identity")
                return
            identity, body = encode(
                self.database.owner,
                self.kind,
                m.target,
                event.revision,
                event.target,
                event.snapshot,
            )
            append_realtime(
                connection,
                "cafe:" + self.database.owner,
                self.kind,
                m.target,
                event.revision,
                identity,
                body,
            )
            outcome = Outcome(aggregateId=m.target, version=event.revision, status="realtime_queued")
            connection.execute(
                """INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome)
                VALUES(%s,%s,%s,%s,%s)""",
                (self.consumer, m.source_id, m.source_hash, target, Jsonb(outcome)),
            )
