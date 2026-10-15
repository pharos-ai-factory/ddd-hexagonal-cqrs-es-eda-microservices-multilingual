"""Explicit browser publication after the preparation fact has been delivered."""

from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository
from operations.contexts.preparation.domain.preparation_ticket import DrinksReadyDomainEvent
from operations.foundation.application import Metadata, Outcome, RealtimeBusPort, RealtimeEvent


class DrinksReadyRealtimeDomainEventHandler:
    """Publish the current ticket view with its own revision, including after delayed delivery."""

    def __init__(self, tickets: TicketReadRepository, realtime: RealtimeBusPort[RealtimeEvent]) -> None:
        self.tickets, self.realtime = tickets, realtime

    def handle(self, metadata: Metadata, event: DrinksReadyDomainEvent) -> Outcome:
        loaded = self.tickets.get(metadata.target)
        if loaded is None or loaded["state"]["orderId"] != event.order_id:
            raise ValueError("Committed preparation ticket is missing or inconsistent")
        self.realtime.publish(
            RealtimeEvent(source=metadata, revision=loaded["version"], snapshot=loaded["state"])
        )
        return Outcome(aggregateId=metadata.target, version=loaded["version"], status="realtime_queued")
