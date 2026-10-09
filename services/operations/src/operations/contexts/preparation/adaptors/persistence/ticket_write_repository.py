"""PostgreSQL aggregate restoration and persistence owned by Preparation."""
from operations.adaptors.mapped_write_repository import MappedWriteRepository
from operations.foundation.write_repository import WriteRepository
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketSnapshot


class PostgresTicketWriteRepository(MappedWriteRepository[TicketSnapshot, PreparationTicket]):
    """Load authoritative aggregates and stage their snapshots in the active transaction."""
    def __init__(self, source: WriteRepository[TicketSnapshot]) -> None:
        super().__init__(source, PreparationTicket.restore, PreparationTicket.snapshot)
