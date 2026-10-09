import pytest
from operations.contexts.collection.adaptors.persistence.pickups import restore, view

ID = "00000000-0000-4000-8000-000000000001"


def test_pickup_read_fields_preserve_code_and_handover_status() -> None:
    state = {"id": ID, "orderId": ID, "customerId": ID, "code": "ABC123", "status": "collected"}
    assert view(restore(state)) == state
    with pytest.raises(ValueError):
        view(restore({**state, "code": None}))
