"""Publish collection's integration contract after its domain event is committed."""

from operations.contexts.collection.domain.pickup import PickupOpenedDomainEvent
from operations.contracts.events.generated.collection import PickupOpenedIntegrationEvent
from operations.foundation.application import IntegrationEventBusPort, Metadata, Outcome


class PickupOpenedDomainEventHandler:
    """Translate a committed private fact through the atomic receipt/publication port."""

    def __init__(self, publications: IntegrationEventBusPort[PickupOpenedIntegrationEvent]) -> None:
        self.publications = publications

    def handle(self, metadata: Metadata, event: PickupOpenedDomainEvent) -> Outcome:
        return self.publications.publish(
            metadata,
            PickupOpenedIntegrationEvent(
                pickupId=event.pickup_id,
                orderId=event.order_id,
                customerId=event.customer_id,
                collectionCode=event.code,
            ),
        )
