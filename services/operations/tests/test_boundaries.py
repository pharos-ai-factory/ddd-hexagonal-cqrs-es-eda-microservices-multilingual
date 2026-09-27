"""External JSON remains untrusted even when internal callers are statically checked."""
import pytest
from operations.adaptors import inputs, receipts
from operations.foundation.domain import CorruptState

ID = "00000000-0000-4000-8000-000000000001"


@pytest.mark.parametrize("value", [None, [], {}, {"code": None}, {"code": 123},
                                 {"code": "ABC123", "extra": True}])
def test_untyped_collection_bodies_cannot_enter_the_application(value: object) -> None:
    with pytest.raises(ValueError):
        inputs.collect_order(value)


@pytest.mark.parametrize("value", [
    {"aggregateId": ID, "version": True, "status": "ready"},
    {"aggregateId": "broken", "version": 1, "status": "ready"},
    {"aggregateId": ID, "version": 1, "status": "ready", "rejection": {"code": "not_found"}},
])
def test_malformed_stored_outcomes_are_retryable_state_failures(value: object) -> None:
    with pytest.raises(CorruptState):
        receipts.outcome(value)
