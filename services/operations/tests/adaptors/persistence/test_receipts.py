import pytest
from operations.adaptors import receipts
from operations.foundation.domain import CorruptState

ID = "00000000-0000-4000-8000-000000000001"

@pytest.mark.parametrize("value", [
    {"aggregateId": ID, "version": True, "status": "ready"},
    {"aggregateId": "broken", "version": 1, "status": "ready"},
    {"aggregateId": ID, "version": 1, "status": "ready", "rejection": {"code": "not_found"}},
])
def test_malformed_stored_outcomes_are_retryable_state_failures(value: object) -> None:
    with pytest.raises(CorruptState):
        receipts.outcome(value)
