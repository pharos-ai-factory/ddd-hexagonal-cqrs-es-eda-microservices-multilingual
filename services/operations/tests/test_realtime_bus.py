"""An explicit publication captures its value and cannot escape its invocation."""

import pytest
from operations.adaptors.realtime.publication_buffer import RealtimeBuffer
from operations.foundation.application import RealtimeEvent


def test_explicit_publication_is_copied_and_scope_is_closed() -> None:
    bus = RealtimeBuffer()
    state = {"status": "queued"}
    assert bus.events == []
    bus.publish(RealtimeEvent(snapshot=state))
    state["status"] = "ready"
    assert bus.events[0].snapshot == {"status": "queued"}
    with pytest.raises(ValueError, match="one realtime payload"):
        bus.publish(RealtimeEvent())
    bus.close()
    with pytest.raises(RuntimeError, match="outside its transaction"):
        bus.publish(RealtimeEvent(snapshot=state))
