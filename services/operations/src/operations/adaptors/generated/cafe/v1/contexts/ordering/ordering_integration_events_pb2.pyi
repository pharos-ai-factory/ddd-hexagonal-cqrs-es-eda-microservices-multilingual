from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Line(_message.Message):
    __slots__ = ("id", "offer_code", "name", "quantity", "minor")
    ID_FIELD_NUMBER: _ClassVar[int]
    OFFER_CODE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    QUANTITY_FIELD_NUMBER: _ClassVar[int]
    MINOR_FIELD_NUMBER: _ClassVar[int]
    id: str
    offer_code: str
    name: str
    quantity: int
    minor: int
    def __init__(self, id: _Optional[str] = ..., offer_code: _Optional[str] = ..., name: _Optional[str] = ..., quantity: _Optional[int] = ..., minor: _Optional[int] = ...) -> None: ...

class OrderPlaced(_message.Message):
    __slots__ = ("order_id", "customer_id", "edition_id", "currency", "lines")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    EDITION_ID_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    LINES_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    customer_id: str
    edition_id: str
    currency: str
    lines: _containers.RepeatedCompositeFieldContainer[Line]
    def __init__(self, order_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., edition_id: _Optional[str] = ..., currency: _Optional[str] = ..., lines: _Optional[_Iterable[_Union[Line, _Mapping]]] = ...) -> None: ...
