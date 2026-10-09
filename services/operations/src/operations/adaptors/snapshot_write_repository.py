"""Transaction-scoped persistence of one authoritative aggregate snapshot."""
from collections.abc import Callable, Mapping
from copy import deepcopy
from psycopg import Connection
from psycopg.types.json import Jsonb
from operations.adaptors.postgres import Row, check_root_identity
from operations.foundation.application import Loaded
from operations.foundation.domain import integer


class PostgresSnapshotWriteRepository[S: Mapping[str, object]]:
    """Own aggregate SQL and stage one root until the command transaction flushes it."""
    def __init__(self, connection: Connection[Row], kind: str, target: str, restore: Callable[[object], S]) -> None:
        self._connection, self._kind, self._target = connection, kind, target
        self._active = True
        row = connection.execute("SELECT version,state FROM cafe.aggregates WHERE kind=%s AND id=%s FOR UPDATE", (kind, target)).fetchone()
        self.version = integer(row["version"]) if row else 0
        self._exists = row is not None
        self._state = restore(row["state"]) if row else None
        if self._state is not None:
            check_root_identity(self._state, target)
        self.pending: S | None = None

    def _check(self) -> None:
        if not self._active:
            raise RuntimeError("Write repository used outside its command transaction")

    def get(self, identity: str) -> Loaded[S] | None:
        self._check()
        if identity != self._target:
            raise ValueError("Command transaction cannot access another aggregate")
        state = self.pending if self.pending is not None else self._state
        return Loaded(exists=True, version=self.version, state=deepcopy(state)) if state is not None else None

    def save(self, state: S) -> None:
        self._check()
        check_root_identity(state, self._target)
        self.pending = deepcopy(state)

    def close(self) -> None:
        self._active = False

    def flush(self) -> int:
        """Called only by the command transaction after successful application execution."""
        if self.pending is None:
            return self.version
        version = self.version + 1
        if self._exists:
            result = self._connection.execute("""UPDATE cafe.aggregates SET version=%s,state=%s
                WHERE kind=%s AND id=%s AND version=%s""", (version, Jsonb(dict(self.pending)), self._kind, self._target, self.version))
            if result.rowcount != 1:
                raise RuntimeError("Optimistic update lost its owner version")
        else:
            self._connection.execute("INSERT INTO cafe.aggregates(kind,id,version,state) VALUES(%s,%s,%s,%s)",
                                     (self._kind, self._target, version, Jsonb(dict(self.pending))))
        return version
