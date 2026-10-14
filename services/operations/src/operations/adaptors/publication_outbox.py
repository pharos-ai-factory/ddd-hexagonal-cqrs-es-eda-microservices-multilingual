"""Atomic receipt and integration outbox persistence for a private-event subscription."""

from collections.abc import Callable
from dataclasses import replace
import re
from psycopg.types.json import Jsonb
from operations.adaptors import codec, receipts
from operations.adaptors.generated.event_registry import PAYLOAD_CODECS
from operations.adaptors.postgres import PostgresContextDatabase
from operations.foundation.application import Metadata, Outcome, Publication
from operations.foundation.domain import identifier


class PostgresPublicationOutbox[P]:
    """Complete one subscription and stage outgoing bytes without changing an aggregate."""

    def __init__(
        self,
        database: PostgresContextDatabase,
        consumer: str,
        kind: str,
        publication: Callable[[P], Publication],
    ) -> None:
        if not consumer.startswith(database.owner + "."):
            raise ValueError("Publication subscription crossed its owner")
        self.database, self.consumer, self.kind = database, consumer, kind
        self.publication = publication

    def publish(self, metadata: Metadata, payload: P) -> Outcome:
        m = metadata
        for value in (m.id, m.target, m.source_id, m.correlation):
            identifier(value)
        if (
            m.consumer != self.consumer
            or m.source_kind != self.kind
            or not 1 <= m.source_version <= 2**53 - 1
            or not m.source_time
            or not re.fullmatch(r"[a-f0-9]{64}", m.source_hash)
        ):
            raise ValueError("Invalid publication receipt metadata")
        target = "publication:" + self.kind + ":" + m.target
        with self.database.pool.connection() as connection, connection.transaction():
            connection.execute(
                "SELECT pg_advisory_xact_lock(hashtextextended(%s,0))",
                ("consumer:" + self.consumer + ":" + m.source_id,),
            )
            previous = connection.execute(
                "SELECT fingerprint,target,outcome FROM cafe.consumer_receipts WHERE consumer=%s AND event_id=%s",
                (self.consumer, m.source_id),
            ).fetchone()
            if previous:
                if previous["fingerprint"] != m.source_hash or previous["target"] != target:
                    raise ValueError("Event identity reused with conflicting bytes or target")
                return receipts.outcome(previous["outcome"], m.target)
            publication = self.publication(payload)
            if publication.name not in PAYLOAD_CODECS:
                raise ValueError("Publication handler must stage an integration event")
            event, body = codec.encode(
                self.database.owner,
                self.kind,
                m.target,
                m.source_version,
                replace(m, id=m.source_id),
                publication,
                occurred_at=m.source_time,
            )
            connection.execute(
                """INSERT INTO cafe.outbox_events(id,event_name,visibility,aggregate_kind,
                aggregate_id,aggregate_version,correlation_id,causation_id,body)
                VALUES(%s,%s,%s,%s,%s,%s,%s,%s,%s)""",
                (
                    event.id,
                    event.name,
                    event.visibility,
                    self.kind,
                    m.target,
                    m.source_version,
                    m.correlation,
                    m.source_id,
                    body,
                ),
            )
            connection.execute("INSERT INTO cafe.dispatches(event_id) VALUES(%s)", (event.id,))
            outcome = Outcome(aggregateId=m.target, version=m.source_version, status="publication_queued")
            connection.execute(
                """INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome)
                VALUES(%s,%s,%s,%s,%s)""",
                (self.consumer, m.source_id, m.source_hash, target, Jsonb(outcome)),
            )
            return outcome
