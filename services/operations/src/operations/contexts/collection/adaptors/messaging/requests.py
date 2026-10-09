"""Explicit mapping between collection wire contracts and application types."""
from google.protobuf.message import Message
from operations.foundation.application import Loaded
from operations.contexts.collection.domain import PickupSnapshot
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_replies_pb2 import Pickup, LoadedPickup, Pickups
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_requests_pb2 import Reply
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommand
from operations.adaptors.generated.cafe.requests.v1.contexts.collection.collection_commands_pb2 import CollectOrder


def collect(value: Message) -> CollectOrderCommand:
    if not isinstance(value, CollectOrder):
        raise ValueError("Unexpected collection payload")
    return CollectOrderCommand(code=value.code)


def item(loaded: Loaded[PickupSnapshot]) -> LoadedPickup:
    s = loaded["state"]
    return LoadedPickup(exists=loaded["exists"], version=loaded["version"], state=Pickup(id=s["id"], order_id=s["orderId"], customer_id=s["customerId"], code=s["code"], status=s["status"]))


def listing(items: list[Loaded[PickupSnapshot]], paged: bool, next_id: str | None) -> Reply:
    result = Pickups(items=[item(value) for value in items], paged=paged)
    if next_id is not None:
        result.next_id = next_id
    return Reply(pickups=result)
