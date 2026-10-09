"""PostgreSQL read repository for Preparation application views."""
from operations.adaptors.postgres import PostgresContextDatabase
from operations.adaptors.snapshot_read_repository import PostgresSnapshotReadRepository
from operations.adaptors.mapped_read_repository import MappedReadRepository
from operations.contexts.preparation.domain.preparation_ticket import TicketSnapshot
from operations.contexts.preparation.application.read_models.ticket import TicketView
from operations.contexts.preparation.adaptors.persistence.ticket_snapshot import restore_ticket_snapshot


def ticket_view(state: TicketSnapshot) -> TicketView:
    return TicketView(id=state["id"], orderId=state["orderId"], customerId=state["customerId"], instructions=state["instructions"], status=state["status"])


class PostgresTicketReadRepository(MappedReadRepository[TicketSnapshot, TicketView]):
    """Expose validated tickets through the Preparation read model."""
    def __init__(self, database: PostgresContextDatabase) -> None:
        super().__init__(PostgresSnapshotReadRepository(database, "ticket", restore_ticket_snapshot), ticket_view)
