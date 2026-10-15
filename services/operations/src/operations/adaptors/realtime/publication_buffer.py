"""Collect explicit publication requests during one application invocation."""

from copy import deepcopy
from operations.foundation.application import RealtimeEvent


class RealtimeBuffer:
    """Stage immutable message values; the owning transaction supplies durable delivery."""

    def __init__(self) -> None:
        self.events: list[RealtimeEvent] = []
        self.active = True

    def publish(self, event: RealtimeEvent) -> None:
        if not self.active:
            raise RuntimeError("Realtime bus used outside its transaction")
        if (event.snapshot is None) == (event.rejection is None):
            raise ValueError("Select one realtime payload")
        self.events.append(deepcopy(event))

    def close(self) -> None:
        self.active = False
