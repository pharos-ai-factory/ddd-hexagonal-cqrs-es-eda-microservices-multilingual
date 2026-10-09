from dataclasses import dataclass
from typing import Protocol
from operations.foundation.application import Loaded, QueryPort


@dataclass(frozen=True)
class PageRequest:
    """Selects a bounded keyset page using the last observed resource identity."""
    limit: int
    after: str | None = None


@dataclass(frozen=True)
class Page[S]:
    """Contains one page of versioned read values and its optional continuation identity."""
    items: list[Loaded[S]]
    next_id: str | None = None


class PagedQueryPort[S](QueryPort[S], Protocol):
    """Extends application reads with an explicit bounded traversal capability."""
    def page(self, request: PageRequest) -> Page[S]: ...
