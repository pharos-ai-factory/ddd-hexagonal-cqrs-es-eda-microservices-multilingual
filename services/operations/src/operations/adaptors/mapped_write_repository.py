"""Translate stored snapshots into aggregates through a scoped write repository."""
from collections.abc import Callable
from operations.foundation.application import Loaded
from operations.foundation.write_repository import WriteRepository


class MappedWriteRepository[S, A]:
    """Keep aggregate restoration and snapshot extraction at the owner boundary."""
    def __init__(self, source: WriteRepository[S], restore: Callable[[S], A], snapshot: Callable[[A], S]) -> None:
        self._source, self._restore, self._snapshot = source, restore, snapshot

    def get(self, identity: str) -> Loaded[A] | None:
        loaded = self._source.get(identity)
        return Loaded(exists=True, version=loaded["version"], state=self._restore(loaded["state"])) if loaded else None

    def save(self, aggregate: A) -> None:
        self._source.save(self._snapshot(aggregate))
