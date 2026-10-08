from operations.adaptors.generated.cafe.v1.contexts.menu import menu_integration_events_pb2 as _menu_integration_events_pb2
from operations.adaptors.generated.cafe.v1.contexts.ordering import ordering_integration_events_pb2 as _ordering_integration_events_pb2
from operations.adaptors.generated.cafe.v1.contexts.preparation import preparation_integration_events_pb2 as _preparation_integration_events_pb2
from operations.adaptors.generated.cafe.v1.contexts.collection import collection_integration_events_pb2 as _collection_integration_events_pb2
from operations.adaptors.generated.cafe.v1.contexts.loyalty import loyalty_integration_events_pb2 as _loyalty_integration_events_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Event(_message.Message):
    __slots__ = ("id", "name", "context", "visibility", "contract_version", "aggregate_kind", "aggregate_id", "aggregate_version", "correlation_id", "causation_id", "occurred_at", "menu_published", "order_placed", "drinks_ready", "pickup_opened", "order_collected", "reward_issued")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    CONTRACT_VERSION_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_KIND_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_ID_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_VERSION_FIELD_NUMBER: _ClassVar[int]
    CORRELATION_ID_FIELD_NUMBER: _ClassVar[int]
    CAUSATION_ID_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    MENU_PUBLISHED_FIELD_NUMBER: _ClassVar[int]
    ORDER_PLACED_FIELD_NUMBER: _ClassVar[int]
    DRINKS_READY_FIELD_NUMBER: _ClassVar[int]
    PICKUP_OPENED_FIELD_NUMBER: _ClassVar[int]
    ORDER_COLLECTED_FIELD_NUMBER: _ClassVar[int]
    REWARD_ISSUED_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    context: str
    visibility: str
    contract_version: int
    aggregate_kind: str
    aggregate_id: str
    aggregate_version: int
    correlation_id: str
    causation_id: str
    occurred_at: str
    menu_published: _menu_integration_events_pb2.MenuPublished
    order_placed: _ordering_integration_events_pb2.OrderPlaced
    drinks_ready: _preparation_integration_events_pb2.DrinksReady
    pickup_opened: _collection_integration_events_pb2.PickupOpened
    order_collected: _collection_integration_events_pb2.OrderCollected
    reward_issued: _loyalty_integration_events_pb2.RewardIssued
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., context: _Optional[str] = ..., visibility: _Optional[str] = ..., contract_version: _Optional[int] = ..., aggregate_kind: _Optional[str] = ..., aggregate_id: _Optional[str] = ..., aggregate_version: _Optional[int] = ..., correlation_id: _Optional[str] = ..., causation_id: _Optional[str] = ..., occurred_at: _Optional[str] = ..., menu_published: _Optional[_Union[_menu_integration_events_pb2.MenuPublished, _Mapping]] = ..., order_placed: _Optional[_Union[_ordering_integration_events_pb2.OrderPlaced, _Mapping]] = ..., drinks_ready: _Optional[_Union[_preparation_integration_events_pb2.DrinksReady, _Mapping]] = ..., pickup_opened: _Optional[_Union[_collection_integration_events_pb2.PickupOpened, _Mapping]] = ..., order_collected: _Optional[_Union[_collection_integration_events_pb2.OrderCollected, _Mapping]] = ..., reward_issued: _Optional[_Union[_loyalty_integration_events_pb2.RewardIssued, _Mapping]] = ...) -> None: ...
