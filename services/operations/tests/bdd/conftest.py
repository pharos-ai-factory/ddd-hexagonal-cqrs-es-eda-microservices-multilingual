"""Scenario-scoped decision probes; durable receipts and delivery use the real-infrastructure lane."""
from copy import deepcopy
from collections.abc import Callable
from dataclasses import dataclass
from pathlib import Path

import pytest

from operations.foundation.application import Change, Loaded, Metadata, Outcome, Publication
from operations.foundation.write_repository import WriteRepository
from operations.adaptors.command_execution import TransactionResult as CommandResult, AggregateTransaction, EventRecordingWriteRepository
from operations.foundation.domain import Rejection, RejectionDetail

SPECIFICATIONS = Path(__file__).resolve().parents[4] / "specifications"
CUSTOMER = "00000000-0000-4000-8000-000000000001"
ORDER = "00000000-0000-4000-8000-000000000002"
ROOT = "00000000-0000-4000-8000-000000000003"


@dataclass
class CommandProbe[S, E]:
    state: S | None = None
    before: S | None = None
    version: int = 0
    previous_version: int = 0
    publications: tuple[Publication, ...] = ()
    rejection: RejectionDetail | None = None
    event: E | None = None

    def execute(self, metadata: Metadata, decide: Callable[[S | None], Change[S]]) -> Outcome:
        self.before, self.previous_version = deepcopy(self.state), self.version
        self.publications, self.rejection = (), None
        try:
            change = decide(deepcopy(self.state))
        except Rejection as error:
            self.rejection = error.outcome()
            return {"aggregateId": metadata.target, "version": self.version, "status": "", "rejection": self.rejection}
        self.publications = change.publications
        if change.changed:
            self.state = deepcopy(change.state)
            self.version += 1
        return {"aggregateId": metadata.target, "version": self.version, "status": change.status}

    def transaction[A](self, restore: Callable[[S], A], snapshot: Callable[[A], S], publications: Callable[[A], tuple[Publication, ...]]) -> AggregateTransaction[A]:
        probe = self
        class ProbeTransaction:
            def execute(self, metadata: Metadata, work: Callable[[WriteRepository[A]], CommandResult]) -> Outcome:
                def decide(state: S | None) -> Change[S]:
                    saved: list[S] = []
                    class Repository:
                        def get(self, identity: str) -> Loaded[A] | None:
                            assert identity == metadata.target
                            return Loaded(exists=True, version=probe.version, state=restore(state)) if state is not None else None
                        def save(self, aggregate: A) -> None:
                            saved.append(snapshot(aggregate))
                    repository = EventRecordingWriteRepository(Repository(), publications)
                    result = work(repository)
                    if not saved:
                        # The placeholder is ignored for a no-op; its type follows the existing snapshot fixture.
                        from typing import cast
                        return Change(cast(S, state), result.status, changed=False, publications=result.publications + repository.publications)
                    return Change(saved[-1], result.status, publications=result.publications + repository.publications)
                return probe.execute(metadata, decide)
        return ProbeTransaction()

    def succeeded(self) -> None:
        assert self.rejection is None, self.rejection

    def unchanged(self) -> None:
        assert self.state == self.before
        assert self.version == self.previous_version
        assert self.publications == ()


    def current(self) -> S:
        assert self.state is not None
        return self.state

    def incoming(self) -> E:
        assert self.event is not None
        return self.event


@pytest.fixture
def metadata() -> Metadata:
    return Metadata(CUSTOMER, ROOT, "scenario-command", CUSTOMER)


def pytest_bdd_apply_tag(tag: str, function: Callable[..., object]) -> bool:
    # The catalogue validates lane/context/identity tags. They are documentation,
    # not arbitrary pytest markers; selection is by feature path in this runner.
    return True
