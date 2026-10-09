"""PostgreSQL restoration and application read mapping owned by Preparation."""
from operations.adaptors.postgres import PostgresContextDatabase, PostgresAggregateQueries
from operations.adaptors.mapped_queries import MappedQueries
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketSnapshot
from operations.contexts.preparation.application.read_models.ticket import TicketView


def restore(value: object) -> TicketSnapshot:
    return PreparationTicket.restore(value).snapshot()


def view(state: TicketSnapshot) -> TicketView:
    return TicketView(id=state["id"], orderId=state["orderId"], customerId=state["customerId"], instructions=state["instructions"], status=state["status"])


class PostgresTicketReader(MappedQueries[TicketSnapshot, TicketView]):
    """Expose validated tickets through the Preparation read model."""
    def __init__(self, database: PostgresContextDatabase) -> None:
        super().__init__(PostgresAggregateQueries(database, "ticket", restore), view)
