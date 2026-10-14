"""Typed boundary definition for an incoming published message."""

from operations.adaptors.generated.message_names import EventName
from collections.abc import Callable
from dataclasses import dataclass

from operations.adaptors.generated.cafe.v1.events_pb2 import Event as WireEvent


@dataclass(frozen=True)
class IncomingEvent[E]:
    """Bind a wire event name to its validated application payload decoder."""

    name: EventName
    parse: Callable[[WireEvent], E]
