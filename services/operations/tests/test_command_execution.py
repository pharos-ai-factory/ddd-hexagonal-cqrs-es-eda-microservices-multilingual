"""Aggregate persistence cannot bypass an independent publication subscription."""

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
        source, lambda _: (Publication("preparation.drinks-ready", {}),)
    )
    with pytest.raises(ValueError, match="private domain"):
        repository.save(1)
    assert not source.saved
    private = EventRecordingWriteRepository[int](
        source, lambda _: (Publication("preparation.drinks-ready.domain", {}),)
    )
    private.save(1)
    assert bool(source.saved)
    assert len(private.publications) == 1
