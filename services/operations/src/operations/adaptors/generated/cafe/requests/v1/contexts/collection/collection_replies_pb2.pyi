from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Pickup(_message.Message):
    __slots__ = ("id", "order_id", "customer_id", "code", "status")
    ID_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    id: str
    order_id: str
    customer_id: str
    code: str
    status: str
    def __init__(self, id: _Optional[str] = ..., order_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., code: _Optional[str] = ..., status: _Optional[str] = ...) -> None: ...

class LoadedPickup(_message.Message):
    __slots__ = ("exists", "version", "state")
    EXISTS_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    exists: bool
    version: int
    state: Pickup
    def __init__(self, exists: _Optional[bool] = ..., version: _Optional[int] = ..., state: _Optional[_Union[Pickup, _Mapping]] = ...) -> None: ...

class Pickups(_message.Message):
    __slots__ = ("items", "paged", "next_id")
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    PAGED_FIELD_NUMBER: _ClassVar[int]
    NEXT_ID_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[LoadedPickup]
    paged: bool
    next_id: str
    def __init__(self, items: _Optional[_Iterable[_Union[LoadedPickup, _Mapping]]] = ..., paged: _Optional[bool] = ..., next_id: _Optional[str] = ...) -> None: ...
