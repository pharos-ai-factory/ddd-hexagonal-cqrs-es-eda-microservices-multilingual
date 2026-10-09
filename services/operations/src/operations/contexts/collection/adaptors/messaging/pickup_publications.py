"""Map private domain facts to the messages this context has chosen to deliver."""
from operations.contexts.collection.domain.pickup import Pickup, PickupOpened, OrderCollected
from operations.contracts import events
from operations.foundation.application import Publication


def pickup_publications(aggregate: Pickup) -> tuple[Publication, ...]:
    """Keep event envelopes and published payloads outside aggregate behaviour."""
    return tuple(Publication("collection.pickup-opened", events.PickupOpened(pickupId=fact.pickup_id, orderId=fact.order_id, customerId=fact.customer_id, collectionCode=fact.code))
        if isinstance(fact, PickupOpened) else Publication("collection.order-collected", events.OrderCollected(orderId=fact.order_id, customerId=fact.customer_id))
        for fact in aggregate.events() if isinstance(fact, (PickupOpened, OrderCollected)))
