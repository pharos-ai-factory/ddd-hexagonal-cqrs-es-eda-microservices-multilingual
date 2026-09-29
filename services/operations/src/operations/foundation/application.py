"""One-aggregate ports; neither transactions nor transport objects cross them."""
from dataclasses import dataclass, field
from collections.abc import Callable, Mapping
from typing import Literal, NotRequired, Protocol, TypedDict
from operations.foundation.domain import Rejection, RejectionDetail


class ApplicationError(Rejection):
    """Expected use-case failure, recorded like a domain rejection."""


class VersionConflictApplicationError(ApplicationError):
    def __init__(self) -> None:
        super().__init__("version_conflict", "The expected aggregate version is stale")


class Outcome(TypedDict):
    aggregateId: str
    version: int
    status: str
    rejection: NotRequired[RejectionDetail]


class Loaded[S](TypedDict):
    exists: Literal[True]
    version: int
    state: S


@dataclass(frozen=True)
class Metadata:
    id: str
    target: str
    name: str
    correlation: str
    expected: int | None = None
    input: Mapping[str, object] = field(default_factory=dict)
    causation: str = ""
    consumer: str = ""
    source_id: str = ""
    source_hash: str = ""


@dataclass(frozen=True)
class Publication:
    name: str
    payload: Mapping[str, object]


@dataclass(frozen=True)
class Change[S]:
    state: S
    status: str
    changed: bool = True
    publications: tuple[Publication, ...] = ()


class CommandPort[S](Protocol):
    def execute(self, metadata: Metadata, decide: Callable[[S | None], Change[S]]) -> Outcome: ...


class QueryPort[S](Protocol):
    def get(self, identity: str) -> Loaded[S] | None: ...
    def list(self) -> list[Loaded[S]]: ...
