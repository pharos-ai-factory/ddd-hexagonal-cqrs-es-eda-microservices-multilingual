"""Central command execution; transaction metadata stays outside feature handlers."""
from collections.abc import Callable
from dataclasses import dataclass
from typing import Protocol
from operations.foundation.application import Loaded, Metadata, Outcome, Publication
from operations.foundation.write_repository import CommandContext, CommandResult, WriteRepository

@dataclass(frozen=True)
class TransactionResult:
    """Internal commit material, including publications from technical fixtures."""
    status: str
    publications: tuple[Publication, ...] = ()

class AggregateTransaction[A](Protocol):
    """Infrastructure boundary for one target in one context database."""
    def execute(self, metadata: Metadata, work: Callable[[WriteRepository[A]], TransactionResult]) -> Outcome: ...

class FeatureCommandHandler[C](Protocol):
    """Plain feature behaviour; each invocation receives its own repository."""
    def execute(self, context: CommandContext, command: C) -> CommandResult: ...

class CommandExecutor[A, C]:
    """Wraps feature execution with the shared local transaction and receipt policy."""
    def __init__(self, transaction: AggregateTransaction[A], handler: Callable[[WriteRepository[A]], FeatureCommandHandler[C]]) -> None:
        self._transaction, self._handler = transaction, handler
    def execute(self, metadata: Metadata, command: C) -> Outcome:
        def work(repository: WriteRepository[A]) -> TransactionResult:
            return TransactionResult(self._handler(repository).execute(CommandContext(metadata.target), command).status)
        return self._transaction.execute(metadata, work)

class EventRecordingWriteRepository[A]:
    """Stages the saved aggregate's pending facts through its owner publication mapper."""
    def __init__(self, source: WriteRepository[A], publications: Callable[[A], tuple[Publication, ...]]) -> None:
        self._source, self._map = source, publications
        self.publications: tuple[Publication, ...] = ()
    def get(self, identity: str) -> Loaded[A] | None:
        return self._source.get(identity)
    def save(self, aggregate: A) -> None:
        self._source.save(aggregate)
        self.publications = self._map(aggregate)
