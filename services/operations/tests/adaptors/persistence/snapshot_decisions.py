"""Technical fixtures drive the real command transaction using snapshot-only decisions."""
from collections.abc import Callable, Mapping
from operations.adaptors.postgres import PostgresContextDatabase
from operations.adaptors.aggregate_transaction import PostgresAggregateTransaction
from operations.foundation.application import Change, Metadata, Outcome
from operations.foundation.write_repository import WriteRepository
from operations.adaptors.command_execution import TransactionResult as CommandResult


class SnapshotDecisionFixture[S: Mapping[str, object]]:
    """Adapt existing storage-failure fixtures without duplicating transaction logic."""
    def __init__(self, database: PostgresContextDatabase, kind: str, restore: Callable[[object], S]) -> None:
        self.transaction = PostgresAggregateTransaction[S, S](database, kind, restore, lambda repository: repository)

    def execute(self, metadata: Metadata, decide: Callable[[S | None], Change[S]]) -> Outcome:
        def work(repository: WriteRepository[S]) -> CommandResult:
            loaded = repository.get(metadata.target)
            change = decide(loaded["state"] if loaded else None)
            if change.changed:
                repository.save(change.state)
            return CommandResult(change.status, change.publications)
        return self.transaction.execute(metadata, work)
