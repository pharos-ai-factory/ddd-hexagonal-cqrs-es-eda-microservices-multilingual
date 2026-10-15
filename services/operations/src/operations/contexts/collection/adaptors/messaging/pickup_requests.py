"""Explicit mapping between collection wire contracts and application types."""

from .generated import request_payloads as inputs


from google.protobuf.message import Message
from operations.foundation.application import Loaded
from operations.contexts.collection.application.read_models.pickup import PickupView
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_replies_pb2 import (
    Pickup,
    LoadedPickup,
    Pickups,
)
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_requests_pb2 import Reply
from operations.contracts.requests.generated.collection import CollectOrderCommand
from operations.contracts.requests.generated.collection import GetPickupQuery
from operations.contracts.requests.generated.collection import ListPickupsQuery
from operations.foundation.domain import identifier
from operations.foundation.pagination import Page


def collect(value: Message) -> CollectOrderCommand:
    return inputs.collect_order(value)


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
    query = inputs.get_pickup(value)
    identifier(query.id)
    return query


def list_pickups(value: Message) -> ListPickupsQuery:
    query = inputs.list_pickups(value)
    if query.page is not None and query.page.after is not None:
        identifier(query.page.after)
    return query


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
