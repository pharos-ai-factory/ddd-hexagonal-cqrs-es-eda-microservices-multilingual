from operations.adaptors.codec import decode
from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.foundation.identity import derived_id
import pytest
from operations.foundation.domain import Rejection

IDS = [f"00000000-0000-4000-8000-{number:012d}" for number in range(1, 5)]

def test_declared_wire_contract_and_owner_tampering() -> None:
    event = Event(id=IDS[0], name="preparation.drinks-ready", context="preparation",
                  visibility="integration", contract_version=1, aggregate_kind="ticket",
                  aggregate_id=IDS[1], aggregate_version=2, correlation_id=IDS[2],
                  causation_id=IDS[3], occurred_at="2026-09-27T12:00:00Z")
    event.drinks_ready.order_id, event.drinks_ready.customer_id = IDS[1], IDS[2]
    raw = event.SerializeToString(deterministic=True)
    decoded, payload = decode(raw)
    assert payload == {"orderId": IDS[1], "customerId": IDS[2]}
    assert decoded.SerializeToString(deterministic=True) == raw
    event.context = "collection"
    with pytest.raises(ValueError, match="owner"):
        decode(event.SerializeToString())
    assert derived_id("earned-grant", IDS[0]) == "f4136e57-6728-8a37-b8f6-a29159aecdc3"


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
