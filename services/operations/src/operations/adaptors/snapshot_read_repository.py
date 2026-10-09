"""Read validated snapshots without opening a command transaction."""
from collections.abc import Callable, Mapping
from operations.adaptors.postgres import PostgresContextDatabase, Row, check_root_identity
from operations.foundation.application import Loaded
from operations.foundation.domain import identifier, integer
from operations.foundation.pagination import Page, PageRequest

class PostgresSnapshotReadRepository[S: Mapping[str, object]]:
    """Restores validated snapshots through the read port without exposing database objects."""
    def __init__(self, database: PostgresContextDatabase, kind: str, restore: Callable[[object], S]) -> None:
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
