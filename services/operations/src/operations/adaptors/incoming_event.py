"""Typed boundary definition for an incoming published message."""
from dataclasses import dataclass
from collections.abc import Callable
from operations.adaptors.generated.cafe.v1.events_pb2 import Event as WireEvent


@dataclass(frozen=True)
class IncomingEvent[E]:
    """Bind a wire event name to its validated application payload decoder."""
    name: str
    parse: Callable[[WireEvent], E]
