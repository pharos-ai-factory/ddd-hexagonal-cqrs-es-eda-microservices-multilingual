from pathlib import Path
from collections.abc import Callable
import pytest
from operations.contexts.preparation.domain import PreparationTicket, TicketState
from operations.contexts.collection.domain import Pickup
from operations.adaptors.codec import decode
from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.foundation.domain import Rejection
from operations.foundation.identity import derived_id

IDS = [f"00000000-0000-4000-8000-{number:012d}" for number in range(1, 5)]


def test_preparation_rejects_skipped_step_without_mutation() -> None:
    ticket = PreparationTicket(TicketState(IDS[0], IDS[1], IDS[2], "2 × coffee"))
    before = ticket.snapshot()
    with pytest.raises(Rejection, match="Only preparing"):
        ticket.complete()
    assert ticket.snapshot() == before
    assert ticket.events() == ()
    ticket.start()
    ticket.complete()
    assert ticket.snapshot()["status"] == "ready"
    assert [fact.name for fact in ticket.events()] == ["PreparationStarted", "DrinksReady"]
    with pytest.raises(Rejection):
        ticket.complete()


def test_collection_code_and_single_handover() -> None:
    pickup = Pickup.open(IDS[0], IDS[1], IDS[2], "AB1234")
    before, facts = pickup.snapshot(), pickup.events()
    with pytest.raises(Rejection) as rejected:
        pickup.collect("WRONG1")
    assert rejected.value.code == "incorrect_collection_code"
    assert (pickup.snapshot(), pickup.events()) == (before, facts)
    pickup.collect("AB1234")
    assert pickup.snapshot()["status"] == "collected"
    with pytest.raises(Rejection):
        pickup.collect("AB1234")
    assert [fact.name for fact in pickup.events()].count("OrderCollected") == 1


def test_restored_root_does_not_reemit_creation() -> None:
    pickup = Pickup.open(IDS[0], IDS[1], IDS[2], "AB1234")
    restored = Pickup.restore(pickup.snapshot())
    assert restored.events() == ()
    snapshot = restored.snapshot()
    snapshot["status"] = "collected"
    assert restored.snapshot()["status"] == "ready"


def test_shared_wire_fixture_and_owner_tampering() -> None:
    root = Path(__file__).resolve().parents[3]
    raw = bytes.fromhex((root/"contracts/events/fixtures/reward-earned.v1.hex").read_text())
    event, grant = decode(raw)
    assert (event.name, event.context, event.visibility) == ("loyalty.reward-earned", "loyalty", "domain")
    assert grant["validDays"] == 7
    assert Event.FromString(raw).SerializeToString(deterministic=True) == raw
    event.context = "preparation"
    with pytest.raises(ValueError, match="owner"):
        decode(event.SerializeToString())
    assert derived_id("earned-grant", IDS[0]) == "f4136e57-6728-8a37-b8f6-a29159aecdc3"


@pytest.mark.parametrize("restore,state", [
    (PreparationTicket.restore, {"id": IDS[0], "orderId": "broken", "customerId": IDS[1],
                                "instructions": "Coffee", "status": "queued"}),
    (Pickup.restore, {"id": IDS[0], "orderId": IDS[1], "customerId": IDS[2],
                      "code": "broken", "status": "ready"}),
    (Pickup.restore, {"id": "broken", "orderId": IDS[1], "customerId": IDS[2],
                      "code": "ABC123", "status": "ready"}),
    (Pickup.restore, {"id": IDS[0], "orderId": IDS[1], "customerId": IDS[2],
                      "code": None, "status": "ready"}),
    (PreparationTicket.restore, {"id": IDS[0], "orderId": IDS[1], "customerId": IDS[2],
                                "instructions": 1, "status": "queued"}),
    (PreparationTicket.restore, {}),
])
def test_corrupt_restoration_is_not_an_expected_business_rejection(
    restore: Callable[[object], object], state: object,
) -> None:
    with pytest.raises(ValueError):
        restore(state)


@pytest.mark.parametrize("missing", ["order_id", "customer_id"])
def test_wire_decoder_rejects_an_omitted_required_business_identity(missing: str) -> None:
    event = Event(id=IDS[0], name="preparation.drinks-ready", context="preparation",
                  visibility="integration", contract_version=1, aggregate_kind="ticket",
                  aggregate_id=IDS[1], aggregate_version=2, correlation_id=IDS[2],
                  causation_id=IDS[3], occurred_at="2026-09-27T12:00:00Z")
    event.drinks_ready.order_id, event.drinks_ready.customer_id = IDS[1], IDS[2]
    event.drinks_ready.ClearField(missing)
    with pytest.raises((ValueError, Rejection)):
        decode(event.SerializeToString())
