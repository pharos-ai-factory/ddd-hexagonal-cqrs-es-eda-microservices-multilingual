"""Both private Pickup facts survive queued-byte recovery with source validation."""

from pathlib import Path
import pytest
from operations.foundation.domain import Rejection
from operations.adaptors.messaging.event_codec import decode
from operations.adaptors.generated.private_event_registry import PRIVATE_CODECS


@pytest.mark.parametrize("name", ["pickup-opened", "order-collected"])
def test_private_collection_fixtures(name: str) -> None:
    path = Path(__file__).parents[3] / (
        "src/operations/contexts/collection/adaptors/messaging/fixtures/" + name + ".domain.hex"
    )
    body = bytes.fromhex(path.read_text().strip())
    event, payload = decode(body)
    codec = PRIVATE_CODECS[event.name]
    assert codec.encode(event, payload) == body
    with pytest.raises(Rejection):
        codec.encode(event, {**payload, "customer_id": "invalid"})
    if name == "pickup-opened":
        with pytest.raises(ValueError, match="identity"):
            codec.encode(event, {**payload, "pickup_id": "55555555-5555-4555-8555-555555555555"})
    event.aggregate_kind = "ticket"
    with pytest.raises(ValueError, match="Foreign"):
        codec.encode(event, payload)
