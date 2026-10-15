"""One-aggregate ports; neither transactions nor transport objects cross them."""

from dataclasses import dataclass, field
from collections.abc import Mapping
from typing import Literal, NotRequired, Protocol, TypedDict
from operations.foundation.domain import Rejection, RejectionDetail


class ApplicationError(Rejection):
    """Expected use-case failure, recorded like a domain rejection."""


class VersionConflictApplicationError(ApplicationError):
    """Records rejection when the caller expected a different aggregate revision."""

    def __init__(self) -> None:
        super().__init__("version_conflict", "The expected aggregate version is stale")


class Outcome(TypedDict):
    """Records the committed aggregate revision or a typed business rejection for stable retries."""

    aggregateId: str
    version: int
    status: str
    rejection: NotRequired[RejectionDetail]


class Loaded[S](TypedDict):
    """Carries an application read value or restored snapshot with its persisted revision."""

    exists: Literal[True]
    version: int
    state: S


@dataclass(frozen=True)
class Metadata:
    """Identifies a command attempt and preserves original material input and source receipt evidence."""

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
    source_kind: str = ""
    source_version: int = 0
    source_time: str = ""


@dataclass(frozen=True)
class Publication:
    """Carries an application-selected event payload for the owner outbox encoder."""

    name: str
    payload: Mapping[str, object]


@dataclass(frozen=True)
class Change[S]:
    """Describes one aggregate transition and its outgoing publications for atomic persistence."""

    state: S
    status: str
    changed: bool = True
    publications: tuple[Publication, ...] = ()


class QueryPort[S](Protocol):
    """Reads restored aggregate snapshots without granting mutation authority."""

    def get(self, identity: str) -> Loaded[S] | None: ...
    def list(self) -> list[Loaded[S]]: ...


class CommandBusPort[C](Protocol):
    """Commit a receiving receipt and outgoing command before returning acceptance."""

    def enqueue(self, metadata: Metadata, command: C) -> Outcome: ...


class IntegrationEventBusPort[P](Protocol):
    """Commit this subscription's receipt and exact outgoing integration event together."""

    def publish(self, metadata: Metadata, event: P) -> Outcome: ...


class QueryBusPort[Q, R](Protocol):
    """Dispatch a typed query to one registered handler and return its result."""

    def ask(self, query: Q) -> R: ...


class DomainEventBusPort[E](Protocol):
    """Stage private domain facts inside their owning transaction."""

    def publish(self, event: E) -> None: ...


class RealtimeBusPort[E](Protocol):
    """Explicitly stage a browser message; the surrounding transaction commits it."""

    def publish(self, event: E) -> None: ...


@dataclass(frozen=True)
class RealtimeEvent:
    """Select a snapshot or committed rejection; owner infrastructure supplies its envelope."""

    source: Metadata | None = None
    revision: int = 0
    snapshot: Mapping[str, object] | None = None
    rejection: RejectionDetail | None = None
