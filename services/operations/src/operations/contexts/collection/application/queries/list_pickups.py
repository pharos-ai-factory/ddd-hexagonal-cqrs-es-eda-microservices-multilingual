from operations.contracts.requests.generated.collection import ListPickupsQuery
from operations.contracts.requests.generated.collection_responses import (
    ListPickupsResponse,
    LoadedPickup,
    Pickups,
)
from operations.contexts.collection.application.ports.pickup_read_repository import PickupReadRepository
from operations.foundation.pagination import Page


class ListPickupsQueryHandler:
    """Preserve revision, missing-resource and pagination policy in a generated response."""

    def __init__(self, read_repository: PickupReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: ListPickupsQuery) -> ListPickupsResponse:
        page = (
            self.read_repository.page(query.page)
            if query.page is not None
            else Page(self.read_repository.list())
        )
        result = Pickups(
            items=[
                LoadedPickup(exists=item["exists"], version=item["version"], state=item["state"])
                for item in page.items
            ],
            paged=query.page is not None,
        )
        if page.next_id is not None:
            result["nextId"] = page.next_id
        return {"pickups": result}
