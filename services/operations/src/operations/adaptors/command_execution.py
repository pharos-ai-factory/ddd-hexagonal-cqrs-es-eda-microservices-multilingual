"""Central command execution; transaction metadata stays outside feature handlers."""

from collections.abc import Callable
from dataclasses import dataclass
from typing import Protocol
from operations.adaptors.generated.private_event_registry import PRIVATE_CODECS
from operations.foundation.application import (
    Loaded,
    Metadata,
    Outcome,
    Publication,
    RealtimeBusPort,
    RealtimeEvent,
    DomainEventBusPort,
)
from operations.foundation.domain import Rejection, RejectionDetail
from operations.adaptors.realtime.publication_buffer import RealtimeBuffer
from operations.foundation.write_repository import CommandContext, CommandResult, WriteRepository


@dataclass(frozen=True)
class TransactionResult:
    """Internal commit material, including publications from technical fixtures."""

    status: str
    publications: tuple[Publication, ...] = ()
    realtime: tuple[RealtimeEvent, ...] = ()
    rejection: RejectionDetail | None = None


class AggregateTransaction[A](Protocol):
    """Infrastructure boundary for one target in one context database."""

    def execute(
        self, metadata: Metadata, work: Callable[[WriteRepository[A]], TransactionResult]
    ) -> Outcome: ...


class FeatureCommandHandler[C](Protocol):
    """Plain feature behaviour; each invocation receives its own repository."""

    def execute(self, context: CommandContext, command: C) -> CommandResult: ...


class CommandExecutor[A, C]:
    """Wraps feature execution with the shared local transaction and receipt policy."""

    def __init__(
        self,
        transaction: AggregateTransaction[A],
        handler: Callable[[WriteRepository[A], RealtimeBusPort[RealtimeEvent]], FeatureCommandHandler[C]],
    ) -> None:
        self._transaction, self._handler = transaction, handler

    def execute(self, metadata: Metadata, command: C) -> Outcome:
        def work(repository: WriteRepository[A]) -> TransactionResult:
            bus = RealtimeBuffer()
            try:
                result = self._handler(repository, bus).execute(
                    CommandContext(metadata.target, metadata.actor_id), command
                )
                return TransactionResult(result.status, realtime=tuple(bus.events))
            except Rejection as rejection:
                return TransactionResult("", realtime=tuple(bus.events), rejection=rejection.outcome())
            finally:
                bus.close()

        return self._transaction.execute(metadata, work)


class EventRecordingWriteRepository[A]:
    """Stages the saved aggregate's pending facts through its owner publication mapper."""

    def __init__(
        self, source: WriteRepository[A], publications: Callable[[A], tuple[Publication, ...]]
    ) -> None:
        self._source, self._map = source, publications
        self.publications: tuple[Publication, ...] = ()

    def get(self, identity: str) -> Loaded[A] | None:
        return self._source.get(identity)

    def save(self, aggregate: A) -> None:
        publications = self._map(aggregate)
        self.publications = ()
        bus: DomainEventBusPort[Publication] = self
        for publication in publications:
            bus.publish(publication)
        self._source.save(aggregate)

    def publish(self, event: Publication) -> None:
        """Stage one owner-private fact for the transaction."""
        if event.name not in PRIVATE_CODECS:
            raise ValueError("Aggregate saves may stage only private domain events")
        self.publications += (event,)
