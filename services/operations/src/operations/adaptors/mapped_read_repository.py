"""Map validated stored snapshots into an application-owned read model."""
from collections.abc import Callable
from operations.foundation.application import Loaded
from operations.foundation.pagination import Page, PageRequest, PagedQueryPort


class MappedReadRepository[S, V]:
    """Preserve revision, missing rows and pagination while translating read values."""
    def __init__(self, source: PagedQueryPort[S], view: Callable[[S], V]) -> None:
        self.source, self.view = source, view

    def _map(self, value: Loaded[S]) -> Loaded[V]:
        return Loaded(exists=True, version=value["version"], state=self.view(value["state"]))

    def get(self, identity: str) -> Loaded[V] | None:
        value = self.source.get(identity)
        return self._map(value) if value is not None else None

    def list(self) -> list[Loaded[V]]:
        return [self._map(value) for value in self.source.list()]

    def page(self, request: PageRequest) -> Page[V]:
        page = self.source.page(request)
        return Page([self._map(value) for value in page.items], page.next_id)
