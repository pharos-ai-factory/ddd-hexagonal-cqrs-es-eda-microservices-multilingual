"""PostgreSQL aggregate restoration and persistence owned by Collection."""
from operations.adaptors.mapped_write_repository import MappedWriteRepository
from operations.foundation.write_repository import WriteRepository
from operations.contexts.collection.domain.pickup import Pickup, PickupSnapshot


class PostgresPickupWriteRepository(MappedWriteRepository[PickupSnapshot, Pickup]):
    """Load authoritative aggregates and stage their snapshots in the active transaction."""
    def __init__(self, source: WriteRepository[PickupSnapshot]) -> None:
        super().__init__(source, Pickup.restore, Pickup.snapshot)
