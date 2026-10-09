from typing import Protocol
from operations.foundation.pagination import PagedQueryPort
from operations.contexts.collection.application.read_models.pickup import PickupView


class PickupReader(PagedQueryPort[PickupView], Protocol):
    """Read pickups with stable revisions and optional keyset pagination."""
