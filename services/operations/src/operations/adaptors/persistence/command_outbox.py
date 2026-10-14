"""Commit receiving evidence and exact owner-command bytes before event acknowledgement."""

from collections.abc import Mapping

from operations.adaptors.messaging.internal_command_codec import InternalCommandCodec
from operations.adaptors.persistence.postgres_database import PostgresContextDatabase
from operations.foundation.application import Metadata, Outcome


class PostgresDurableCommandOutbox[C: Mapping[str, object]]:
    """Return acceptance only after the receiving receipt and exact command bytes commit."""

    def __init__(self, database: PostgresContextDatabase, codec: InternalCommandCodec[C]) -> None:
        if database.owner != codec.owner:
            raise ValueError("Command outbox crossed its owner")
        self.database, self.codec = database, codec

    def enqueue(self, metadata: Metadata, command: C) -> Outcome:
        body = self.codec.encode(metadata, command)
        with self.database.pool.connection() as connection:
            with connection.transaction():
                connection.execute(
                    "SELECT pg_advisory_xact_lock(hashtextextended(%s,0))",
                    ("handoff:" + metadata.consumer + ":" + metadata.source_id,),
                )
                previous = connection.execute(
                    """SELECT id,fingerprint,target FROM cafe.internal_commands
                    WHERE consumer=%s AND event_id=%s""",
                    (metadata.consumer, metadata.source_id),
                ).fetchone()
                if previous:
                    if (
                        str(previous["id"]) != metadata.id
                        or previous["fingerprint"] != metadata.source_hash
                        or str(previous["target"]) != metadata.target
                    ):
                        raise ValueError("Conflicting event hand-off identity")
                else:
                    connection.execute(
                        """INSERT INTO cafe.internal_commands(id,consumer,event_id,fingerprint,target,body)
                        VALUES(%s,%s,%s,%s,%s,%s)""",
                        (
                            metadata.id,
                            metadata.consumer,
                            metadata.source_id,
                            metadata.source_hash,
                            metadata.target,
                            body,
                        ),
                    )
                    connection.execute(
                        "INSERT INTO cafe.internal_command_dispatches(event_id) VALUES(%s)", (metadata.id,)
                    )
        return Outcome(aggregateId=metadata.target, version=0, status="queued")
