"""Versioned wire mapping. Generated messages never enter domain packages."""
from datetime import datetime, timezone
from collections.abc import Mapping
import json
from pathlib import Path
from google.protobuf.json_format import MessageToDict, ParseDict
from google.protobuf.message import Message
from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.adaptors.generated.cafe.realtime.v1.realtime_pb2 import Publication as Realtime
from operations.contracts.events import DrinksReady, OrderLine, OrderPlaced
from operations.foundation.application import Metadata, Publication
from operations.foundation.domain import identifier, record
from operations.foundation.identity import new_id

CATALOGUE = {item["name"]: item for item in json.loads(
    (Path(__file__).parent/"generated/catalogue.json").read_text())}
PAYLOADS = {
    "menu.drink-published": "drink_published", "menu.edition-published": "menu_published",
    "ordering.order-placed": "order_placed", "preparation.drinks-ready": "drinks_ready",
    "collection.pickup-opened": "pickup_opened", "collection.order-collected": "order_collected",
    "loyalty.reward-earned": "reward_earned", "loyalty.reward-issued": "reward_issued",
    "communication.notification-requested": "notification_requested",
}
SOURCE_IDENTITIES = {
    "menu.drink-published": "drink_id", "menu.edition-published": "edition_id",
    "ordering.order-placed": "order_id", "collection.pickup-opened": "pickup_id",
    "loyalty.reward-earned": "account_id", "loyalty.reward-issued": "reward_id",
    "communication.notification-requested": "notification_id",
}


def validate(event: Event) -> None:
    definition = CATALOGUE.get(event.name)
    if not definition or (event.context, event.visibility, event.aggregate_kind) != (
        definition["owner"], definition["visibility"], definition["kind"]
    ) or event.contract_version != 1:
        raise ValueError("Unknown event contract or owner")
    for value in (event.id, event.aggregate_id, event.correlation_id, event.causation_id):
        identifier(value)
    if event.aggregate_version < 1 or event.WhichOneof("payload") != PAYLOADS[event.name]:
        raise ValueError("Missing or mismatched event payload")
    datetime.fromisoformat(event.occurred_at.replace("Z", "+00:00"))
    payload: Message = getattr(event, PAYLOADS[event.name])
    # Proto3 omits default values from ListFields; required IDs must be checked even when absent.
    for field in payload.DESCRIPTOR.fields:
        if field.name.endswith("_id"):
            identifier(getattr(payload, field.name))
    source = SOURCE_IDENTITIES.get(event.name)
    if source and getattr(payload, source) != event.aggregate_id:
        raise ValueError("Payload does not identify its source aggregate")
    if event.name == "ordering.order-placed":
        order = event.order_placed
        count = 0
        seen = set()
        for line in order.lines:
            identifier(line.id)
            if line.id in seen or not line.name or not 1 <= line.quantity <= 5 or line.minor < 0:
                raise ValueError("Invalid accepted order line")
            seen.add(line.id)
            count += line.quantity
        if not 1 <= count <= 5 or order.order_id != event.aggregate_id:
            raise ValueError("Invalid accepted order")


def decode(body: bytes) -> tuple[Event, dict[str, object]]:
    if len(body) > 256*1024:
        raise ValueError("Event exceeds its wire limit")
    event = Event.FromString(body)
    validate(event)
    payload = record(MessageToDict(getattr(event, PAYLOADS[event.name])))
    return event, payload


def order_placed(event: Event) -> OrderPlaced:
    if event.name != "ordering.order-placed":
        raise ValueError("Expected OrderPlaced")
    order = event.order_placed
    return OrderPlaced(orderId=order.order_id, customerId=order.customer_id, editionId=order.edition_id,
        currency=order.currency, lines=[OrderLine(id=line.id, offerCode=line.offer_code, name=line.name,
            quantity=line.quantity, minor=line.minor) for line in order.lines])


def drinks_ready(event: Event) -> DrinksReady:
    if event.name != "preparation.drinks-ready":
        raise ValueError("Expected DrinksReady")
    return DrinksReady(orderId=event.drinks_ready.order_id, customerId=event.drinks_ready.customer_id)


def encode(owner: str, kind: str, identity: str, version: int, metadata: Metadata,
           publication: Publication) -> tuple[Event, bytes]:
    definition = CATALOGUE[publication.name]
    event = Event(id=new_id(), name=publication.name, context=owner,
                  visibility=definition["visibility"], contract_version=1,
                  aggregate_kind=kind, aggregate_id=identity, aggregate_version=version,
                  correlation_id=metadata.correlation, causation_id=metadata.id,
                  occurred_at=datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"))
    ParseDict(dict(publication.payload), getattr(event, PAYLOADS[publication.name]))
    validate(event)
    return event, event.SerializeToString(deterministic=True)


def realtime(owner: str, kind: str, identity: str, version: int, state: Mapping[str, object]) -> tuple[str, bytes]:
    event = Realtime(event_id=new_id(), contract_version=1, context=owner,
                     aggregate_kind=kind, aggregate_id=identity, revision=version)
    ParseDict(dict(state), getattr(event, kind))
    return event.event_id, event.SerializeToString(deterministic=True)
