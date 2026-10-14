"""Stage Preparation's private facts in the aggregate transaction."""

from operations.contexts.preparation.domain.preparation_ticket import (
    PreparationTicket,
    DrinksReadyDomainEvent,
)
from operations.foundation.application import Publication


def ticket_publications(aggregate: PreparationTicket) -> tuple[Publication, ...]:
    """Select the private fact consumed by the durable integration publication handler."""
    return tuple(
        Publication(
            "preparation.drinks-ready.domain",
            {"orderId": fact.order_id, "customerId": fact.customer_id},
        )
        for fact in aggregate.events()
        if isinstance(fact, DrinksReadyDomainEvent)
    )
