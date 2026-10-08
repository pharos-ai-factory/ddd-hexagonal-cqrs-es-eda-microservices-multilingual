from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Ticket(_message.Message):
    __slots__ = ("id", "order_id", "customer_id", "instructions", "status")
    ID_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    INSTRUCTIONS_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    id: str
    order_id: str
    customer_id: str
    instructions: str
    status: str
    def __init__(self, id: _Optional[str] = ..., order_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., instructions: _Optional[str] = ..., status: _Optional[str] = ...) -> None: ...

class LoadedTicket(_message.Message):
    __slots__ = ("exists", "version", "state")
    EXISTS_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    exists: bool
    version: int
    state: Ticket
    def __init__(self, exists: _Optional[bool] = ..., version: _Optional[int] = ..., state: _Optional[_Union[Ticket, _Mapping]] = ...) -> None: ...

class Tickets(_message.Message):
    __slots__ = ("items", "paged", "next_id")
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    PAGED_FIELD_NUMBER: _ClassVar[int]
    NEXT_ID_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[LoadedTicket]
    paged: bool
    next_id: str
    def __init__(self, items: _Optional[_Iterable[_Union[LoadedTicket, _Mapping]]] = ..., paged: _Optional[bool] = ..., next_id: _Optional[str] = ...) -> None: ...
