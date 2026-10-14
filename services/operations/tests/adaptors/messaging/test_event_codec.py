from operations.adaptors.codec import decode
from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.foundation.identity import derived_id
from operations.adaptors.generated.cafe.v1.contexts.preparation.preparation_integration_events_pb2 import (
    DrinksReady,
)
import pytest
from operations.foundation.domain import Rejection

IDS = [f"00000000-0000-4000-8000-{number:012d}" for number in range(1, 5)]


def test_declared_wire_contract_and_owner_tampering() -> None:
    event = Event(
        id=IDS[0],
        name="preparation.drinks-ready",
        context="preparation",
        visibility="integration",
        contract_version=1,
        aggregate_kind="ticket",
        aggregate_id=IDS[1],
        aggregate_version=2,
        correlation_id=IDS[2],
        causation_id=IDS[3],
        occurred_at="2026-09-27T12:00:00Z",
    )
    payload = DrinksReady(order_id=IDS[1], customer_id=IDS[2])
    event.payload = payload.SerializeToString()
    raw = event.SerializeToString(deterministic=True)
    decoded, data = decode(raw)
    assert data == {"orderId": IDS[1], "customerId": IDS[2]}
    assert decoded.SerializeToString(deterministic=True) == raw
    event.context = "collection"
    with pytest.raises(ValueError, match="owner"):
        decode(event.SerializeToString())
    assert derived_id("earned-grant", IDS[0]) == "f4136e57-6728-8a37-b8f6-a29159aecdc3"


@pytest.mark.parametrize("missing", ["order_id", "customer_id"])
def test_wire_decoder_rejects_an_omitted_required_business_identity(missing: str) -> None:
    event = Event(
        id=IDS[0],
        name="preparation.drinks-ready",
        context="preparation",
        visibility="integration",
        contract_version=1,
        aggregate_kind="ticket",
        aggregate_id=IDS[1],
        aggregate_version=2,
        correlation_id=IDS[2],
        causation_id=IDS[3],
        occurred_at="2026-09-27T12:00:00Z",
    )
    payload = DrinksReady(order_id=IDS[1], customer_id=IDS[2])
    event.payload = payload.SerializeToString()
    payload.ClearField(missing)
    event.payload = payload.SerializeToString()
    with pytest.raises((ValueError, Rejection)):
        decode(event.SerializeToString())


@pytest.mark.parametrize(
    "name,version,revision",
    [
        ("preparation.unknown", 1, 1),
        ("preparation.drinks-ready", 2, 1),
        ("preparation.drinks-ready", 1, 2**53),
    ],
)
def test_undeclared_contracts_and_nonportable_revisions_fail(name: str, version: int, revision: int) -> None:
    event = Event(
        id=IDS[0],
        name=name,
        context="preparation",
        visibility="integration",
        contract_version=version,
        aggregate_kind="ticket",
        aggregate_id=IDS[1],
        aggregate_version=revision,
        correlation_id=IDS[2],
        causation_id=IDS[3],
        occurred_at="2026-09-27T12:00:00Z",
        payload=DrinksReady(order_id=IDS[1], customer_id=IDS[2]).SerializeToString(),
    )
    with pytest.raises(ValueError):
        decode(event.SerializeToString())
