"""An explicit publication captures its value and cannot escape its invocation."""

import pytest
from operations.adaptors.realtime.publication_buffer import RealtimeBuffer
from operations.foundation.application import RealtimeMessage, context_target, snapshot_message


def test_explicit_publication_is_copied_and_scope_is_closed() -> None:
    bus = RealtimeBuffer()
    state = {"status": "queued"}
    assert bus.events == []
    bus.publish(snapshot_message(context_target("preparation"), state))
    state["status"] = "ready"
    assert bus.events[0].snapshot == {"status": "queued"}
    with pytest.raises(ValueError, match="one realtime payload"):
        bus.publish(RealtimeMessage(context_target("preparation")))
    bus.close()
    with pytest.raises(RuntimeError, match="outside its transaction"):
        bus.publish(snapshot_message(context_target("preparation"), state))
