"""Explicit mapping between collection wire contracts and application types."""

from google.protobuf.message import Message
from operations.foundation.application import Loaded
from operations.contexts.collection.application.read_models.pickup import PickupView
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_replies_pb2 import (
    Pickup,
    LoadedPickup,
    Pickups,
)
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_requests_pb2 import Reply
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommand
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_commands_pb2 import (
    CollectOrder,
)
from operations.contexts.collection.application.queries.get_pickup import GetPickupQuery
from operations.contexts.collection.application.queries.list_pickups import ListPickupsQuery
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_queries_pb2 import (
    GetPickup,
    ListPickups,
)
from operations.foundation.domain import identifier
from operations.foundation.pagination import Page, PageRequest


def collect(value: Message) -> CollectOrderCommand:
    if not isinstance(value, CollectOrder):
        raise ValueError("Unexpected collection payload")
    return CollectOrderCommand(code=value.code)


def item(loaded: Loaded[PickupView]) -> LoadedPickup:
    s = loaded["state"]
    return LoadedPickup(
        exists=loaded["exists"],
        version=loaded["version"],
        state=Pickup(
            id=s["id"], order_id=s["orderId"], customer_id=s["customerId"], code=s["code"], status=s["status"]
        ),
    )


def listing(items: list[Loaded[PickupView]], paged: bool, next_id: str | None) -> Reply:
    result = Pickups(items=[item(value) for value in items], paged=paged)
    if next_id is not None:
        result.next_id = next_id
    return Reply(pickups=result)


def get_pickup(value: Message) -> GetPickupQuery:
    """Decode the published item query into its application input."""
    if not isinstance(value, GetPickup):
        raise ValueError("Unexpected get pickup payload")
    return GetPickupQuery(identifier(value.id))


def list_pickups(value: Message) -> ListPickupsQuery:
    """Preserve complete reads and explicitly requested pagination."""
    if not isinstance(value, ListPickups):
        raise ValueError("Unexpected list pickups payload")
    page = value.page if value.HasField("page") else None
    return ListPickupsQuery(
        PageRequest(page.limit, identifier(page.after) if page.HasField("after") else None)
        if page is not None
        else None
    )


def get_reply(query: GetPickupQuery, loaded: Loaded[PickupView] | None) -> Reply:
    """Map a missing resource or its current revision to the owner reply."""
    if loaded is None:
        reply = Reply()
        reply.error.code = "not_found"
        return reply
    return Reply(pickup=item(loaded))


def list_reply(query: ListPickupsQuery, page: Page[PickupView]) -> Reply:
    """Retain the pagination flag and continuation from the application result."""
    return listing(page.items, query.page is not None, page.next_id)
