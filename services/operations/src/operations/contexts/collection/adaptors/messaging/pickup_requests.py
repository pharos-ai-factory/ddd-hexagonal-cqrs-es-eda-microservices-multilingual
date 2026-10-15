"""Explicit mapping between collection wire contracts and application types."""

from .generated import request_payloads as inputs


from google.protobuf.message import Message
from operations.contracts.requests.generated.collection import CollectOrderCommand
from operations.contracts.requests.generated.collection import GetPickupQuery
from operations.contracts.requests.generated.collection import ListPickupsQuery
from operations.foundation.domain import identifier


def collect(value: Message) -> CollectOrderCommand:
    return inputs.collect_order(value)


def get_pickup(value: Message) -> GetPickupQuery:
    query = inputs.get_pickup(value)
    identifier(query.id)
    return query


def list_pickups(value: Message) -> ListPickupsQuery:
    query = inputs.list_pickups(value)
    if query.page is not None and query.page.after is not None:
        identifier(query.page.after)
    return query
