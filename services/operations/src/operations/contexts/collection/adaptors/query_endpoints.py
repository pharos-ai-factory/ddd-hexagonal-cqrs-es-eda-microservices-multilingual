"""Translate transport query arguments into named Collection use cases."""
from operations.foundation.application import Loaded
from operations.foundation.pagination import Page, PageRequest
from operations.contexts.collection.application.read_models.pickup import PickupView
from operations.contexts.collection.application.queries.get_pickup import GetPickupQuery, GetPickupQueryHandler
from operations.contexts.collection.application.queries.list_pickups import ListPickupsQuery, ListPickupsQueryHandler


class PickupQueryEndpoints:
    """Translate RabbitMQ query arguments into named application use cases."""
    def __init__(self, get: GetPickupQueryHandler, listing: ListPickupsQueryHandler) -> None:
        self.get_handler, self.list_handler = get, listing

    def get(self, identity: str) -> Loaded[PickupView] | None:
        return self.get_handler.execute(GetPickupQuery(identity))

    def list(self) -> list[Loaded[PickupView]]:
        return self.list_handler.execute(ListPickupsQuery()).items

    def page(self, request: PageRequest) -> Page[PickupView]:
        return self.list_handler.execute(ListPickupsQuery(request))
