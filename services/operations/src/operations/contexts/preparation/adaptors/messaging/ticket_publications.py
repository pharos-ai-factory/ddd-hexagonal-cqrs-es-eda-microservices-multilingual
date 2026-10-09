"""Map private domain facts to the messages this context has chosen to deliver."""
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, DrinksReady
from operations.contracts import events
from operations.foundation.application import Publication


def ticket_publications(aggregate: PreparationTicket) -> tuple[Publication, ...]:
    """Keep event envelopes and published payloads outside aggregate behaviour."""
    return tuple(Publication("preparation.drinks-ready", events.DrinksReady(orderId=fact.order_id, customerId=fact.customer_id))
        for fact in aggregate.events() if isinstance(fact, DrinksReady))
