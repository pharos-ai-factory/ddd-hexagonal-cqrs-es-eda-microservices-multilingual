"""Aggregate persistence cannot bypass an independent publication subscription."""

from operations.adaptors.generated.message_names import PreparationDomainEvent, PreparationIntegrationEvent

import pytest
from operations.adaptors.command_execution import EventRecordingWriteRepository
from operations.foundation.application import Loaded, Publication


class RepositoryProbe:
    def __init__(self) -> None:
        self.saved = False

    def get(self, identity: str) -> Loaded[int] | None:
        return None

    def save(self, aggregate: int) -> None:
        self.saved = True


def test_aggregate_save_rejects_integration_publication() -> None:
    source = RepositoryProbe()
    repository = EventRecordingWriteRepository[int](
        source, lambda _: (Publication(PreparationIntegrationEvent.DRINKS_READY, {}),)
    )
    with pytest.raises(ValueError, match="private domain"):
        repository.save(1)
    assert not source.saved
    private = EventRecordingWriteRepository[int](
        source, lambda _: (Publication(PreparationDomainEvent.DRINKS_READY, {}),)
    )
    private.save(1)
    assert bool(source.saved)
    assert len(private.publications) == 1


def test_handler_receives_authenticated_actor() -> None:
    from collections.abc import Callable
    from operations.adaptors.command_execution import CommandExecutor, TransactionResult
    from operations.foundation.application import Metadata, Outcome
    from operations.foundation.write_repository import CommandContext, CommandResult, WriteRepository

    class Transaction:
        def execute(
            self, metadata: Metadata, work: Callable[[WriteRepository[int]], TransactionResult]
        ) -> Outcome:
            result = work(RepositoryProbe())
            return Outcome(aggregateId=metadata.target, version=1, status=result.status)

    class Handler:
        def execute(self, context: CommandContext, command: int) -> CommandResult:
            assert context.target == "target"
            assert context.actor_id == actor
            return CommandResult("accepted")

    for actor in ("", "11111111-1111-4111-8111-111111111111"):
        executor = CommandExecutor[int, int](Transaction(), lambda repository, bus: Handler())
        assert (
            executor.execute(Metadata("id", "target", "command", "correlation", actor_id=actor), 0)["status"]
            == "accepted"
        )
