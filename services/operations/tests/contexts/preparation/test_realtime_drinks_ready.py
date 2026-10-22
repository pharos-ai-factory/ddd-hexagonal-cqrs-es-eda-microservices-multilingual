"""Delayed facts refresh the current owner view with the view's own revision."""

import pytest
from operations.adaptors.realtime.publication_buffer import RealtimeBuffer
from operations.contexts.preparation.application.event_handlers.realtime_drinks_ready import (
    DrinksReadyRealtimeDomainEventHandler,
)
from operations.contexts.preparation.application.read_models.ticket import TicketView
from operations.contexts.preparation.domain.preparation_ticket import DrinksReadyDomainEvent
from operations.foundation.application import Loaded, Metadata, context_target
from operations.foundation.pagination import Page, PageRequest

ID = "00000000-0000-4000-8000-000000000001"
ORDER = "00000000-0000-4000-8000-000000000002"


class TicketReader:
    def __init__(self) -> None:
        self.loaded: Loaded[TicketView] | None = {
            "exists": True,
            "version": 4,
            "state": {"id": ID, "orderId": ORDER, "customerId": ID, "instructions": "Tea", "status": "ready"},
        }

    def get(self, identity: str) -> Loaded[TicketView] | None:
        assert identity == ID
        return self.loaded

    def list(self) -> list[Loaded[TicketView]]:
        raise AssertionError("Unexpected list")

    def page(self, request: PageRequest) -> Page[TicketView]:
        raise AssertionError("Unexpected page")


def test_delayed_fact_uses_read_revision_and_requires_committed_view() -> None:
    reader, bus = TicketReader(), RealtimeBuffer()
    handler = DrinksReadyRealtimeDomainEventHandler(reader, bus)
    source = Metadata(ID, ID, "drinks-ready", ID, source_version=2)
    event = DrinksReadyDomainEvent(ORDER, ID)
    assert handler.handle(source, event)["version"] == 4
    assert bus.events[0].revision == 4
    assert bus.events[0].source == source
    assert bus.events[0].target == context_target("preparation")
    reader.loaded = None
    with pytest.raises(ValueError, match="missing or inconsistent"):
        handler.handle(source, event)
    assert len(bus.events) == 1
