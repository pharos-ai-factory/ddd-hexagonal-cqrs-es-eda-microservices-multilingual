from operations.contexts.collection.domain.pickup import Pickup, PickupAlreadyCollectedDomainError
import pytest
from operations.foundation.domain import Rejection

IDS = [f"00000000-0000-4000-8000-{number:012d}" for number in range(1, 5)]

def test_collection_code_and_single_handover() -> None:
    pickup = Pickup.open(IDS[0], IDS[1], IDS[2], "AB1234")
    before, facts = pickup.snapshot(), pickup.events()
    with pytest.raises(Rejection) as rejected:
        pickup.collect("WRONG1")
    assert rejected.value.code == "incorrect_collection_code"
    assert (pickup.snapshot(), pickup.events()) == (before, facts)
    pickup.collect("AB1234")
    assert pickup.snapshot()["status"] == "collected"
    with pytest.raises(PickupAlreadyCollectedDomainError) as repeated:
        pickup.collect("AB1234")
    assert repeated.value.outcome() == {
        "code": "pickup_already_collected", "message": "This pickup has already been collected"}
    assert [fact.name for fact in pickup.events()].count("OrderCollected") == 1


def test_restored_root_does_not_reemit_creation() -> None:
    pickup = Pickup.open(IDS[0], IDS[1], IDS[2], "AB1234")
    restored = Pickup.restore(pickup.snapshot())
    assert restored.events() == ()
    snapshot = restored.snapshot()
    snapshot["status"] = "collected"
    assert restored.snapshot()["status"] == "ready"


@pytest.mark.parametrize("state", [{"id": IDS[0], "orderId": IDS[1], "customerId": IDS[2], "code": "broken", "status": "ready"},
 {"id": "broken", "orderId": IDS[1], "customerId": IDS[2], "code": "ABC123", "status": "ready"},
 {"id": IDS[0], "orderId": IDS[1], "customerId": IDS[2], "code": None, "status": "ready"}])
def test_corrupt_restoration_is_retryable(state: object) -> None:
    with pytest.raises(ValueError):
        Pickup.restore(state)
