"""Validate persisted pickup state for both command and query adaptors."""
from operations.contexts.collection.domain.pickup import Pickup, PickupSnapshot


def restore_pickup_snapshot(value: object) -> PickupSnapshot:
    """Rehydrate the owner aggregate and return its validated storage snapshot."""
    return Pickup.restore(value).snapshot()
