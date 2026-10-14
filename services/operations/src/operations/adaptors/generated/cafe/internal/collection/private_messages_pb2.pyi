from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PickupOpenedDomainEvent(_message.Message):
    __slots__ = ("pickup_id", "order_id", "customer_id", "collection_code")
    PICKUP_ID_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_CODE_FIELD_NUMBER: _ClassVar[int]
    pickup_id: str
    order_id: str
    customer_id: str
    collection_code: str
    def __init__(self, pickup_id: _Optional[str] = ..., order_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., collection_code: _Optional[str] = ...) -> None: ...

class OrderCollectedDomainEvent(_message.Message):
    __slots__ = ("order_id", "customer_id")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    customer_id: str
    def __init__(self, order_id: _Optional[str] = ..., customer_id: _Optional[str] = ...) -> None: ...

class PrivateEvent(_message.Message):
    __slots__ = ("id", "name", "context", "visibility", "format_revision", "aggregate_kind", "aggregate_id", "aggregate_version", "correlation_id", "causation_id", "occurred_at", "pickup_opened", "order_collected")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    FORMAT_REVISION_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_KIND_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_ID_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_VERSION_FIELD_NUMBER: _ClassVar[int]
    CORRELATION_ID_FIELD_NUMBER: _ClassVar[int]
    CAUSATION_ID_FIELD_NUMBER: _ClassVar[int]
    OCCURRED_AT_FIELD_NUMBER: _ClassVar[int]
    PICKUP_OPENED_FIELD_NUMBER: _ClassVar[int]
    ORDER_COLLECTED_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    context: str
    visibility: str
    format_revision: int
    aggregate_kind: str
    aggregate_id: str
    aggregate_version: int
    correlation_id: str
    causation_id: str
    occurred_at: str
    pickup_opened: PickupOpenedDomainEvent
    order_collected: OrderCollectedDomainEvent
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., context: _Optional[str] = ..., visibility: _Optional[str] = ..., format_revision: _Optional[int] = ..., aggregate_kind: _Optional[str] = ..., aggregate_id: _Optional[str] = ..., aggregate_version: _Optional[int] = ..., correlation_id: _Optional[str] = ..., causation_id: _Optional[str] = ..., occurred_at: _Optional[str] = ..., pickup_opened: _Optional[_Union[PickupOpenedDomainEvent, _Mapping]] = ..., order_collected: _Optional[_Union[OrderCollectedDomainEvent, _Mapping]] = ...) -> None: ...
