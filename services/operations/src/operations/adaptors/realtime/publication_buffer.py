"""Collect explicit publication requests during one application invocation."""

from copy import deepcopy
from operations.foundation.application import RealtimeMessage


class RealtimeBuffer:
    """Stage immutable message values; the owning transaction supplies durable delivery."""

    def __init__(self) -> None:
        self.events: list[RealtimeMessage] = []
        self.active = True

    def publish(self, event: RealtimeMessage) -> None:
        if not self.active:
            raise RuntimeError("Realtime bus used outside its transaction")
        if (event.snapshot is None) == (event.rejection is None):
            raise ValueError("Select one realtime payload")
        if not event.target.id or event.target.kind not in ("context", "principal"):
            raise ValueError("Select one realtime delivery target")
        self.events.append(deepcopy(event))

    def close(self) -> None:
        self.active = False
