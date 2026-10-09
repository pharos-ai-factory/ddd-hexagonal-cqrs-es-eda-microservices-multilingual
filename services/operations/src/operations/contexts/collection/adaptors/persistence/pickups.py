"""PostgreSQL restoration and application read mapping owned by Collection."""
from operations.adaptors.postgres import PostgresContextDatabase, PostgresAggregateQueries
from operations.adaptors.mapped_queries import MappedQueries
from operations.contexts.collection.domain.pickup import Pickup, PickupSnapshot
from operations.contexts.collection.application.read_models.pickup import PickupView


def restore(value: object) -> PickupSnapshot:
    return Pickup.restore(value).snapshot()


def view(state: PickupSnapshot) -> PickupView:
    return PickupView(id=state["id"], orderId=state["orderId"], customerId=state["customerId"], code=state["code"], status=state["status"])


class PostgresPickupReader(MappedQueries[PickupSnapshot, PickupView]):
    """Expose validated pickups through the Collection read model."""
    def __init__(self, database: PostgresContextDatabase) -> None:
        super().__init__(PostgresAggregateQueries(database, "pickup", restore), view)
