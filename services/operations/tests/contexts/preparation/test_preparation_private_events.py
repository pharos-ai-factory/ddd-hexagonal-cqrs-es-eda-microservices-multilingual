"""Queued private event bytes retain their format and owner boundary."""

from pathlib import Path
import pytest
from operations.foundation.domain import Rejection
from operations.adaptors.codec import decode
from operations.adaptors.generated.private_event_registry import PRIVATE_CODECS
from operations.contexts.preparation.adaptors.messaging.private_events import DRINKS_READY_DOMAIN
from operations.contexts.preparation.domain.preparation_ticket import DrinksReadyDomainEvent


def test_private_drinks_ready_fixture_and_scope() -> None:
    path = (
        Path(__file__).parents[3]
        / "src/operations/contexts/preparation/adaptors/messaging/fixtures/drinks-ready.domain.hex"
    )
    body = bytes.fromhex(path.read_text().strip())
    event, payload = decode(body)
    codec = PRIVATE_CODECS[event.name]
    assert codec.encode(event, payload) == body
    assert isinstance(DRINKS_READY_DOMAIN.parse(event), DrinksReadyDomainEvent)
    event.context = "collection"
    with pytest.raises(ValueError, match="Foreign"):
        codec.encode(event, payload)
    event.context = "preparation"
    event.contract_version = 2
    with pytest.raises(ValueError):
        codec.encode(event, payload)
    event.contract_version = 1
    with pytest.raises(Rejection):
        codec.encode(event, {**payload, "order_id": "invalid"})
