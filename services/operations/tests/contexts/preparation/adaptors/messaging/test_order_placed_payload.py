"""Preparation rejects inconsistent immutable order evidence at its boundary."""

from operations.adaptors.generated.message_names import OrderingIntegrationEvent

import pytest
from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.adaptors.generated.cafe.v1.contexts.ordering.ordering_integration_events_pb2 import (
    OrderPlaced,
)
from operations.contexts.preparation.adaptors.messaging.order_placed_payload import order_placed

IDS = [f"00000000-0000-4000-8000-{number:012d}" for number in range(1, 7)]


def event(quantities: list[int], duplicate: bool = False) -> Event:
    payload = OrderPlaced(order_id=IDS[0], customer_id=IDS[1], edition_id=IDS[2], currency="EUR")
    for index, quantity in enumerate(quantities):
        payload.lines.add(
            id=IDS[3 if duplicate else 3 + index], offer_code="TEA", name="Tea", quantity=quantity, minor=0
        )
    return Event(
        name=OrderingIntegrationEvent.ORDER_PLACED, aggregate_id=IDS[0], payload=payload.SerializeToString()
    )


def test_free_lines_preserve_zero_amount_and_total_quantity() -> None:
    result = order_placed(event([2, 3]))
    assert result["lines"][0]["minor"] == 0
    assert [line["quantity"] for line in result["lines"]] == [2, 3]


@pytest.mark.parametrize("quantities,duplicate", [([3, 3], False), ([1, 1], True)])
def test_inconsistent_order_evidence_is_rejected(quantities: list[int], duplicate: bool) -> None:
    with pytest.raises(ValueError, match="evidence"):
        order_placed(event(quantities, duplicate))
