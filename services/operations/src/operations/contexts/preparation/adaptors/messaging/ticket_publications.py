"""Map private domain facts to the messages this context has chosen to deliver."""

from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, DrinksReady
from operations.contracts.events import preparation as events
from operations.foundation.application import Publication
from operations.adaptors.generated.event_preparation import publish_DrinksReady


def ticket_publications(aggregate: PreparationTicket) -> tuple[Publication, ...]:
    """Keep event envelopes and published payloads outside aggregate behaviour."""
    return tuple(
        publish_DrinksReady(events.DrinksReady(orderId=fact.order_id, customerId=fact.customer_id))
        for fact in aggregate.events()
        if isinstance(fact, DrinksReady)
    )
