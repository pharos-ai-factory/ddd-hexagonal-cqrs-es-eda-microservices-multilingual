"""Map private domain facts to the messages this context has chosen to deliver."""

from operations.contexts.collection.domain.pickup import Pickup, PickupOpened, OrderCollected
from operations.contracts.events.generated import collection as events
from operations.foundation.application import Publication
from operations.adaptors.generated.event_collection import publish_PickupOpened, publish_OrderCollected


def pickup_publications(aggregate: Pickup) -> tuple[Publication, ...]:
    """Keep event envelopes and published payloads outside aggregate behaviour."""
    return tuple(
        publish_PickupOpened(
            events.PickupOpened(
                pickupId=fact.pickup_id,
                orderId=fact.order_id,
                customerId=fact.customer_id,
                collectionCode=fact.code,
            )
        )
        if isinstance(fact, PickupOpened)
        else publish_OrderCollected(events.OrderCollected(orderId=fact.order_id, customerId=fact.customer_id))
        for fact in aggregate.events()
        if isinstance(fact, (PickupOpened, OrderCollected))
    )
