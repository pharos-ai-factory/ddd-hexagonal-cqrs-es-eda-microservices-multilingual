from operations.contracts.requests.generated.collection import GetPickupQuery
from operations.contracts.requests.generated.collection_responses import GetPickupResponse, LoadedPickup
from operations.contexts.collection.application.ports.pickup_read_repository import PickupReadRepository


class GetPickupQueryHandler:
    """Preserve revision, missing-resource and pagination policy in a generated response."""

    def __init__(self, read_repository: PickupReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: GetPickupQuery) -> GetPickupResponse:
        value = self.read_repository.get(query.id)
        if value is None:
            return {"error": {"code": "not_found", "message": ""}}
        return {
            "pickup": LoadedPickup(exists=value["exists"], version=value["version"], state=value["state"])
        }
