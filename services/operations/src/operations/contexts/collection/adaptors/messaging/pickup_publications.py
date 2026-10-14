"""Stage Collection's private facts in the aggregate transaction."""

from operations.adaptors.generated.message_names import CollectionDomainEvent

from operations.contexts.collection.domain.pickup import Pickup, PickupOpenedDomainEvent
from operations.foundation.application import Publication


def pickup_publications(aggregate: Pickup) -> tuple[Publication, ...]:
    """Select durable private reactions without constructing published contracts."""
    return tuple(
        Publication(
            CollectionDomainEvent.PICKUP_OPENED,
            {
                "pickupId": fact.pickup_id,
                "orderId": fact.order_id,
                "customerId": fact.customer_id,
                "collectionCode": fact.code,
            },
        )
        if isinstance(fact, PickupOpenedDomainEvent)
        else Publication(
            CollectionDomainEvent.ORDER_COLLECTED,
            {
                "orderId": fact.order_id,
                "customerId": fact.customer_id,
            },
        )
        for fact in aggregate.events()
    )
