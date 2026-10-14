"""Publish preparation's integration contract after its domain event is committed."""

from operations.contexts.preparation.domain.preparation_ticket import DrinksReadyDomainEvent
from operations.contracts.events.generated.preparation import DrinksReadyIntegrationEvent
from operations.foundation.application import DurablePublicationPort, Metadata, Outcome


class DrinksReadyDomainEventHandler:
    """Translate a committed private fact through the atomic receipt/publication port."""

    def __init__(self, publications: DurablePublicationPort[DrinksReadyIntegrationEvent]) -> None:
        self.publications = publications

    def handle(self, metadata: Metadata, event: DrinksReadyDomainEvent) -> Outcome:
        return self.publications.publish(
            metadata,
            DrinksReadyIntegrationEvent(
                orderId=event.order_id,
                customerId=event.customer_id,
            ),
        )
