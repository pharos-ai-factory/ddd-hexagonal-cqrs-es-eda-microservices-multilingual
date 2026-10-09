from dataclasses import dataclass
from operations.foundation.application import Loaded
from operations.contexts.collection.application.ports.pickup_read_repository import PickupReadRepository
from operations.contexts.collection.application.read_models.pickup import PickupView


@dataclass(frozen=True)
class GetPickupQuery:
    """Select one pickup."""
    identity: str


class GetPickupQueryHandler:
    """Execute the GetPickup read use case through its owning application port."""
    def __init__(self, read_repository: PickupReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: GetPickupQuery) -> Loaded[PickupView] | None:
        return self.read_repository.get(query.identity)
