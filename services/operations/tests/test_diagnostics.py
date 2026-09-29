import json
import logging
import pytest
from operations.adaptors.diagnostics import failure, process_workers


def test_failure_before_claim_is_retained_and_secrets_are_redacted(caplog: pytest.LogCaptureFixture) -> None:
    with caplog.at_level(logging.WARNING):
        failure("test", "connection", RuntimeError("amqp://user:secret@host"))
        failure("test", "consumer", ValueError("secret body"), "event-id", "correlation-id", True)
    state = process_workers()
    assert state["test/connection"]["failures"] == 1
    assert state["test/consumer"]["deadLetterTransfers"] == 1
    assert "secret" not in json.dumps(state)
    assert "secret" not in caplog.text
    state.clear()
    assert process_workers()["test/connection"]["failures"] == 1
