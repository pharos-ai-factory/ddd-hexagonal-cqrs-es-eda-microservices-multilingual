from dataclasses import dataclass
from operations.foundation.pagination import Page, PageRequest
from operations.contexts.collection.application.ports.pickup_read_repository import PickupReadRepository
from operations.contexts.collection.application.read_models.pickup import PickupView


@dataclass(frozen=True)
class ListPickupsQuery:
    """Select pickups or an explicitly requested page."""
    page: PageRequest | None = None


class ListPickupsQueryHandler:
    """Execute the ListPickups read use case through its owning application port."""
    def __init__(self, read_repository: PickupReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: ListPickupsQuery) -> Page[PickupView]:
        return self.read_repository.page(query.page) if query.page is not None else Page(self.read_repository.list())
