"""Context-owned PostgreSQL connection pool and storage identity helpers."""
from hashlib import sha256
import json
from pathlib import Path
from collections.abc import Mapping
from psycopg import Connection
from psycopg.rows import dict_row
from psycopg_pool import ConnectionPool
from operations.foundation.domain import CorruptState

type Row = dict[str, object]

CONTEXT_SCHEMAS = json.loads((Path(__file__).parent/"generated/context-persistence.json").read_text())
SCHEMA = json.loads((Path(__file__).parent/"generated/persistence.json").read_text())

def fingerprint(value: object) -> str:
    return sha256(json.dumps(value, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def check_root_identity(state: Mapping[str, object], target: str) -> None:
    if state.get("id") != target:
        raise CorruptState("Snapshot identity differs from its storage key")


class PostgresContextDatabase:
    """Owns one context pool and verifies its database identity, grants and migration ledger."""
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
