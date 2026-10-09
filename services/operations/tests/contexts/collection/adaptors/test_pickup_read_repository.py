import pytest
from operations.contexts.collection.adaptors.persistence.pickup_snapshot import restore_pickup_snapshot
from operations.contexts.collection.adaptors.persistence.pickup_read_repository import pickup_view

ID = "00000000-0000-4000-8000-000000000001"


def test_pickup_read_fields_preserve_code_and_handover_status() -> None:
    state = {"id": ID, "orderId": ID, "customerId": ID, "code": "ABC123", "status": "collected"}
    assert pickup_view(restore_pickup_snapshot(state)) == state
    with pytest.raises(ValueError):
        pickup_view(restore_pickup_snapshot({**state, "code": None}))
