"""Coordinate one repository transaction with receipts, outcomes and outgoing intent."""
from collections.abc import Callable, Mapping
from dataclasses import replace
from psycopg import Connection
from psycopg.types.json import Jsonb
from operations.adaptors.postgres import PostgresContextDatabase, Row, fingerprint
from operations.adaptors.snapshot_write_repository import PostgresSnapshotWriteRepository
from operations.adaptors import codec, receipts
from operations.adaptors.replies import append as append_reply
from operations.foundation.application import Metadata, Outcome, Publication, VersionConflictApplicationError
from operations.foundation.domain import Rejection, identifier
from operations.foundation.write_repository import WriteRepository
from operations.adaptors.command_execution import TransactionResult, AggregateTransaction, EventRecordingWriteRepository


class PostgresAggregateTransaction[S: Mapping[str, object], A](AggregateTransaction[A]):
    """Commits one aggregate, command outcome, receipts and outgoing event/realtime intent atomically."""
    def __init__(self, database: PostgresContextDatabase, kind: str, restore: Callable[[object], S], repository: Callable[[WriteRepository[S]], WriteRepository[A]], publications: Callable[[A], tuple[Publication, ...]] = lambda _: ()) -> None:
        self.database, self.kind = database, kind
        self.restore, self.repository, self.publications = restore, repository, publications

    def execute(self, metadata: Metadata, work: Callable[[WriteRepository[A]], TransactionResult]) -> Outcome:
        m = metadata
        for value in (m.id, m.target, m.correlation):
            identifier(value)
        target = self.kind+":"+m.target
        digest = fingerprint({"expected": m.expected, "input": dict(m.input)})
        with self.database.pool.connection() as connection, connection.transaction():
            connection.execute("SELECT pg_advisory_xact_lock(hashtextextended(%s,0))", (target,))
            connection.execute("SELECT set_config('cafe.command_target',%s,true)", (target,))
            if m.consumer:
                connection.execute("SELECT pg_advisory_xact_lock(hashtextextended(%s,0))",
                                   ("consumer:"+m.consumer+":"+m.source_id,))
                previous = connection.execute("""SELECT fingerprint,target,outcome FROM cafe.consumer_receipts
                    WHERE consumer=%s AND event_id=%s""", (m.consumer, m.source_id)).fetchone()
                if previous:
                    if previous["fingerprint"] != m.source_hash or previous["target"] != target:
                        raise ValueError("Event identity reused with conflicting bytes or target")
                    return append_reply(connection, receipts.outcome(previous["outcome"], m.target))
            receipt = connection.execute("""SELECT fingerprint,outcome FROM cafe.command_receipts
                WHERE kind=%s AND aggregate_id=%s AND command_name=%s AND command_id=%s""",
                (self.kind, m.target, m.name, m.id)).fetchone()
            if receipt:
                if receipt["fingerprint"] != digest:
                    return append_reply(connection, {"aggregateId": m.target, "version": 0, "status": "", "rejection": {
                        "code": "idempotency_conflict", "message": "The command identity has different input"}})
                outcome = receipts.outcome(receipt["outcome"], m.target)
                self._incoming(connection, m, target, outcome)
                return append_reply(connection, outcome)
            repository = PostgresSnapshotWriteRepository(connection, self.kind, m.target, self.restore)
            version = repository.version
            outcome = {"aggregateId": m.target, "version": version, "status": ""}
            try:
                if m.expected is not None and m.expected != version:
                    raise VersionConflictApplicationError()
                try:
                    owner_repository = EventRecordingWriteRepository(self.repository(repository), self.publications)
                    result = work(owner_repository)
                    result = replace(result, publications=result.publications + owner_repository.publications)
                finally:
                    repository.close()
                outcome["status"] = result.status
            except Rejection as rejection:
                repository.close()
                outcome["rejection"] = rejection.outcome()
            else:
                outcome["version"] = self._persist_result(connection, m, repository, result)
            connection.execute("""INSERT INTO cafe.command_receipts(kind,aggregate_id,command_name,command_id,
                fingerprint,outcome,correlation_id,causation_id) VALUES(%s,%s,%s,%s,%s,%s,%s,%s)""",
                (self.kind, m.target, m.name, m.id, digest, Jsonb(outcome), m.correlation, m.causation or None))
            self._incoming(connection, m, target, outcome)
            return append_reply(connection, outcome)

    def _persist_result(self, connection: Connection[Row], m: Metadata,
                        repository: PostgresSnapshotWriteRepository[S], result: TransactionResult) -> int:
        if repository.pending is None:
            if result.publications:
                raise ValueError("A no-op cannot emit new publications")
            return repository.version
        version = repository.flush()
        for publication in result.publications:
            event, body = codec.encode(self.database.owner, self.kind, m.target, version, m, publication)
            connection.execute("""INSERT INTO cafe.outbox_events(id,event_name,visibility,aggregate_kind,
                aggregate_id,aggregate_version,correlation_id,causation_id,body)
                VALUES(%s,%s,%s,%s,%s,%s,%s,%s,%s)""", (event.id, event.name, event.visibility,
                self.kind, m.target, version, m.correlation, m.id, body))
            connection.execute("INSERT INTO cafe.dispatches(event_id) VALUES(%s)", (event.id,))
        event_id, body = codec.realtime(self.database.owner, self.kind, m.target, version, repository.pending)
        connection.execute("""INSERT INTO cafe.realtime_publications
            (id,channel,aggregate_kind,aggregate_id,revision,body) VALUES(%s,%s,%s,%s,%s,%s)""",
            (event_id, "cafe:"+self.database.owner, self.kind, m.target, version, body))
        connection.execute("INSERT INTO cafe.realtime_dispatches(event_id) VALUES(%s)", (event_id,))
        return version

    @staticmethod
    def _incoming(connection: Connection[Row], m: Metadata, target: str, outcome: Outcome) -> None:
        if m.consumer:
            connection.execute("""INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome)
                VALUES(%s,%s,%s,%s,%s)""", (m.consumer, m.source_id, m.source_hash, target, Jsonb(outcome)))
