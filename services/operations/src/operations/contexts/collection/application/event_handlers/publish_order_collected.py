"""Publish collection's integration contract after its domain event is committed."""

from operations.contexts.collection.domain.pickup import OrderCollectedDomainEvent
from operations.contracts.events.generated.collection import OrderCollectedIntegrationEvent
from operations.foundation.application import IntegrationEventBusPort, Metadata, Outcome


class OrderCollectedDomainEventHandler:
    """Translate a committed private fact through the atomic receipt/publication port."""

    def __init__(self, publications: IntegrationEventBusPort[OrderCollectedIntegrationEvent]) -> None:
        self.publications = publications

    def handle(self, metadata: Metadata, event: OrderCollectedDomainEvent) -> Outcome:
        return self.publications.publish(
            metadata,
            OrderCollectedIntegrationEvent(
                orderId=event.order_id,
                customerId=event.customer_id,
            ),
        )
