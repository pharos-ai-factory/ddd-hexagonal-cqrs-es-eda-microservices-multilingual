from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Selection(_message.Message):
    __slots__ = ("offer_code", "name", "minor")
    OFFER_CODE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    MINOR_FIELD_NUMBER: _ClassVar[int]
    offer_code: str
    name: str
    minor: int
    def __init__(self, offer_code: _Optional[str] = ..., name: _Optional[str] = ..., minor: _Optional[int] = ...) -> None: ...

class Line(_message.Message):
    __slots__ = ("id", "selection", "quantity")
    ID_FIELD_NUMBER: _ClassVar[int]
    SELECTION_FIELD_NUMBER: _ClassVar[int]
    QUANTITY_FIELD_NUMBER: _ClassVar[int]
    id: str
    selection: Selection
    quantity: int
    def __init__(self, id: _Optional[str] = ..., selection: _Optional[_Union[Selection, _Mapping]] = ..., quantity: _Optional[int] = ...) -> None: ...

class Order(_message.Message):
    __slots__ = ("id", "customer_id", "edition_id", "currency", "status", "lines")
    ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    EDITION_ID_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    LINES_FIELD_NUMBER: _ClassVar[int]
    id: str
    customer_id: str
    edition_id: str
    currency: str
    status: str
    lines: _containers.RepeatedCompositeFieldContainer[Line]
    def __init__(self, id: _Optional[str] = ..., customer_id: _Optional[str] = ..., edition_id: _Optional[str] = ..., currency: _Optional[str] = ..., status: _Optional[str] = ..., lines: _Optional[_Iterable[_Union[Line, _Mapping]]] = ...) -> None: ...
