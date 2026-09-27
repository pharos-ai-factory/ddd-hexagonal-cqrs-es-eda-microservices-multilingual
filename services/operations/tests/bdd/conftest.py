"""Scenario-scoped decision probes; durable receipts and delivery use the real-infrastructure lane."""
from copy import deepcopy
from collections.abc import Callable
from dataclasses import dataclass
from pathlib import Path

import pytest

from operations.foundation.application import Change, Metadata, Outcome, Publication
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
