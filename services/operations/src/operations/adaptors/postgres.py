"""The transaction boundary commits one root, receipts and both delivery intents."""
from hashlib import sha256
import json
from pathlib import Path
from collections.abc import Callable, Mapping
from psycopg import Connection
from psycopg.rows import dict_row
from psycopg.types.json import Jsonb
from psycopg_pool import ConnectionPool
from operations.foundation.pagination import Page, PageRequest
from operations.foundation.application import Change, Loaded, Metadata, Outcome, VersionConflictApplicationError
from operations.foundation.domain import CorruptState, Rejection, identifier, integer
from operations.adaptors import codec, receipts
from operations.adaptors.replies import append as append_reply

type Row = dict[str, object]

CONTEXT_SCHEMAS = json.loads((Path(__file__).parent/"generated/context-persistence.json").read_text())
SCHEMA = json.loads((Path(__file__).parent/"generated/persistence.json").read_text())

def fingerprint(value: object) -> str:
    return sha256(json.dumps(value, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def check_root_identity(state: Mapping[str, object], target: str) -> None:
    if state.get("id") != target:
        raise CorruptState("Snapshot identity differs from its storage key")


class Database:
    def __init__(self, owner: str, url: str) -> None:
        self.owner = owner
        self.pool: ConnectionPool[Connection[Row]] = ConnectionPool(url, min_size=1, max_size=4, open=True,
                                   kwargs={"autocommit": True, "row_factory": dict_row})
        try:
            with self.pool.connection() as connection:
                row = connection.execute("""SELECT current_database() AS database,
                    rolsuper OR rolcreatedb OR rolcreaterole
                    OR has_database_privilege(current_user,current_database(),'CREATE')
                    OR has_schema_privilege(current_user,'cafe','CREATE') AS privileged
                    FROM pg_roles WHERE rolname=current_user""").fetchone()
                if not row or row["database"] != "cafe_"+owner or row["privileged"]:
                    raise RuntimeError("Database owner or runtime privilege mismatch")
                for version, table in ((1, "schema_version"), (2, "schema_migrations")):
                    row = connection.execute(f"SELECT checksum FROM cafe.{table} WHERE version=%s", (version,)).fetchone()
                    if not row or row["checksum"] != SCHEMA[str(version)]:
                        raise RuntimeError("Database migration checksum mismatch")
                row = connection.execute("SELECT owner FROM cafe.context_identity WHERE singleton").fetchone()
                if not row or row["owner"] != owner or owner not in CONTEXT_SCHEMAS:
                    raise RuntimeError("Context schema identity mismatch")
                migrations = connection.execute("SELECT version,checksum FROM cafe.context_migrations").fetchall()
                actual = {str(item["version"]): item["checksum"] for item in migrations}
                if actual != CONTEXT_SCHEMAS[owner]:
                    raise RuntimeError("Context migration checksum mismatch")
                connection.execute("SELECT id FROM cafe.realtime_publications LIMIT 0")
        except Exception:
            self.pool.close()
            raise


class Commands[S: Mapping[str, object]]:
    def __init__(self, database: Database, kind: str, restore: Callable[[object], S]) -> None:
        self.database, self.kind = database, kind
        self.restore = restore

    def execute(self, metadata: Metadata, decide: Callable[[S | None], Change[S]]) -> Outcome:
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
            loaded = connection.execute("SELECT version,state FROM cafe.aggregates WHERE kind=%s AND id=%s FOR UPDATE",
                                        (self.kind, m.target)).fetchone()
            version = integer(loaded["version"]) if loaded else 0
            state = self.restore(loaded["state"]) if loaded else None
            if state is not None:
                check_root_identity(state, m.target)
            outcome = {"aggregateId": m.target, "version": version, "status": ""}
            try:
                if m.expected is not None and m.expected != version:
                    raise VersionConflictApplicationError()
                change = decide(state)
                outcome["status"] = change.status
            except Rejection as rejection:
                outcome["rejection"] = rejection.outcome()
            else:
                outcome["version"] = self._persist_change(connection, m, loaded is not None, version, change)
            connection.execute("""INSERT INTO cafe.command_receipts(kind,aggregate_id,command_name,command_id,
                fingerprint,outcome,correlation_id,causation_id) VALUES(%s,%s,%s,%s,%s,%s,%s,%s)""",
                (self.kind, m.target, m.name, m.id, digest, Jsonb(outcome), m.correlation, m.causation or None))
            self._incoming(connection, m, target, outcome)
            return append_reply(connection, outcome)

    def _persist_change(self, connection: Connection[Row], m: Metadata, exists: bool,
                        version: int, change: Change[S]) -> int:
        if not change.changed:
            if change.publications:
                raise ValueError("A no-op cannot emit new publications")
            return version
        check_root_identity(change.state, m.target)
        version += 1
        if exists:
            connection.execute("""UPDATE cafe.aggregates SET version=%s,state=%s
                WHERE kind=%s AND id=%s""", (version, Jsonb(dict(change.state)), self.kind, m.target))
        else:
            connection.execute("INSERT INTO cafe.aggregates(kind,id,version,state) VALUES(%s,%s,%s,%s)",
                               (self.kind, m.target, version, Jsonb(dict(change.state))))
        for publication in change.publications:
            event, body = codec.encode(self.database.owner, self.kind, m.target, version, m, publication)
            connection.execute("""INSERT INTO cafe.outbox_events(id,event_name,visibility,aggregate_kind,
                aggregate_id,aggregate_version,correlation_id,causation_id,body)
                VALUES(%s,%s,%s,%s,%s,%s,%s,%s,%s)""", (event.id, event.name, event.visibility,
                self.kind, m.target, version, m.correlation, m.id, body))
            connection.execute("INSERT INTO cafe.dispatches(event_id) VALUES(%s)", (event.id,))
        event_id, body = codec.realtime(self.database.owner, self.kind, m.target, version, change.state)
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


class Queries[S: Mapping[str, object]]:
    def __init__(self, database: Database, kind: str, restore: Callable[[object], S]) -> None:
        self.database, self.kind = database, kind
        self.restore = restore

    def get(self, identity: str) -> Loaded[S] | None:
        identifier(identity)
        with self.database.pool.connection() as connection:
            row = connection.execute("SELECT id,version,state FROM cafe.aggregates WHERE kind=%s AND id=%s",
                                     (self.kind, identity)).fetchone()
            return self._loaded(row) if row else None

    def list(self) -> list[Loaded[S]]:
        with self.database.pool.connection() as connection:
            rows = connection.execute("SELECT id,version,state FROM cafe.aggregates WHERE kind=%s ORDER BY id",
                                      (self.kind,)).fetchall()
            return [self._loaded(row) for row in rows]

    def page(self, request: PageRequest) -> Page[S]:
        if request.limit < 1 or request.limit > 100:
            raise ValueError("Invalid page size")
        if request.after:
            identifier(request.after)
        with self.database.pool.connection() as connection:
            rows = connection.execute(
                "SELECT id,version,state FROM cafe.aggregates WHERE kind=%s "
                "AND (%s::uuid IS NULL OR id>%s::uuid) ORDER BY id LIMIT %s",
                (self.kind, request.after, request.after, request.limit + 1)).fetchall()
            selected = rows[:request.limit]
            return Page([self._loaded(row) for row in selected],
                        str(selected[-1]["id"]) if len(rows) > request.limit else None)

    def _loaded(self, row: Row) -> Loaded[S]:
        state = self.restore(row["state"])
        check_root_identity(state, str(row["id"]))
        return {"exists": True, "version": integer(row["version"]), "state": state}
