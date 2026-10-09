"""Single-aggregate persistence capabilities with an explicit transaction boundary."""
from dataclasses import dataclass
from typing import Protocol
from operations.foundation.application import Loaded


@dataclass(frozen=True)
class CommandResult:
    """Business status returned by a feature command handler."""
    status: str


class WriteRepository[A](Protocol):
    """Load and save the targeted aggregate within the active command transaction."""
    def get(self, identity: str) -> Loaded[A] | None: ...
    def save(self, aggregate: A) -> None: ...


@dataclass(frozen=True)
class CommandContext:
    """Identifies the aggregate targeted by this business operation."""
    target: str
