from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Offer(_message.Message):
    __slots__ = ("code", "drink_id", "drink_revision", "name", "minor", "currency")
    CODE_FIELD_NUMBER: _ClassVar[int]
    DRINK_ID_FIELD_NUMBER: _ClassVar[int]
    DRINK_REVISION_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    MINOR_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    code: str
    drink_id: str
    drink_revision: int
    name: str
    minor: int
    currency: str
    def __init__(self, code: _Optional[str] = ..., drink_id: _Optional[str] = ..., drink_revision: _Optional[int] = ..., name: _Optional[str] = ..., minor: _Optional[int] = ..., currency: _Optional[str] = ...) -> None: ...

class MenuPublished(_message.Message):
    __slots__ = ("edition_id", "currency", "offers")
    EDITION_ID_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    OFFERS_FIELD_NUMBER: _ClassVar[int]
    edition_id: str
    currency: str
    offers: _containers.RepeatedCompositeFieldContainer[Offer]
    def __init__(self, edition_id: _Optional[str] = ..., currency: _Optional[str] = ..., offers: _Optional[_Iterable[_Union[Offer, _Mapping]]] = ...) -> None: ...
