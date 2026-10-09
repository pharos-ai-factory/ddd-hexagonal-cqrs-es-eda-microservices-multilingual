"""PostgreSQL read repository for Collection application views."""
from operations.adaptors.postgres import PostgresContextDatabase
from operations.adaptors.snapshot_read_repository import PostgresSnapshotReadRepository
from operations.adaptors.mapped_read_repository import MappedReadRepository
from operations.contexts.collection.domain.pickup import PickupSnapshot
from operations.contexts.collection.application.read_models.pickup import PickupView
from operations.contexts.collection.adaptors.persistence.pickup_snapshot import restore_pickup_snapshot


def pickup_view(state: PickupSnapshot) -> PickupView:
    return PickupView(id=state["id"], orderId=state["orderId"], customerId=state["customerId"], code=state["code"], status=state["status"])


class PostgresPickupReadRepository(MappedReadRepository[PickupSnapshot, PickupView]):
    """Expose validated pickups through the Collection read model."""
    def __init__(self, database: PostgresContextDatabase) -> None:
        super().__init__(PostgresSnapshotReadRepository(database, "pickup", restore_pickup_snapshot), pickup_view)
