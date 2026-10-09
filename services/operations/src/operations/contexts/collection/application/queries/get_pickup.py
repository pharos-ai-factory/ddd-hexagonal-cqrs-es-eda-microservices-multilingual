from dataclasses import dataclass
from operations.foundation.application import Loaded
from operations.contexts.collection.application.ports.pickup_reader import PickupReader
from operations.contexts.collection.application.read_models.pickup import PickupView


@dataclass(frozen=True)
class GetPickupQuery:
    """Select one pickup."""
    identity: str


class GetPickupQueryHandler:
    """Execute the GetPickup read use case through its owning application port."""
    def __init__(self, reader: PickupReader) -> None:
        self.reader = reader

    def execute(self, query: GetPickupQuery) -> Loaded[PickupView] | None:
        return self.reader.get(query.identity)
