from dataclasses import dataclass
from typing import Protocol
from operations.foundation.application import Loaded, QueryPort


@dataclass(frozen=True)
class PageRequest:
    limit: int
    after: str | None = None


@dataclass(frozen=True)
class Page[S]:
    items: list[Loaded[S]]
    next_id: str | None = None


class PagedQueryPort[S](QueryPort[S], Protocol):
    def page(self, request: PageRequest) -> Page[S]: ...
