from google.protobuf import descriptor_pb2 as _descriptor_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor
EVENT_FIELD_NUMBER: _ClassVar[int]
event: _descriptor.FieldDescriptor
RULES_FIELD_NUMBER: _ClassVar[int]
rules: _descriptor.FieldDescriptor

class EventContract(_message.Message):
    __slots__ = ("name", "owner", "kind", "version")
    NAME_FIELD_NUMBER: _ClassVar[int]
    OWNER_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    name: str
    owner: str
    kind: str
    version: int
    def __init__(self, name: _Optional[str] = ..., owner: _Optional[str] = ..., kind: _Optional[str] = ..., version: _Optional[int] = ...) -> None: ...

class FieldRules(_message.Message):
    __slots__ = ("uuid", "non_empty", "source_identity", "timestamp", "pattern", "minimum", "maximum", "min_items", "max_items")
    UUID_FIELD_NUMBER: _ClassVar[int]
    NON_EMPTY_FIELD_NUMBER: _ClassVar[int]
    SOURCE_IDENTITY_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    PATTERN_FIELD_NUMBER: _ClassVar[int]
    MINIMUM_FIELD_NUMBER: _ClassVar[int]
    MAXIMUM_FIELD_NUMBER: _ClassVar[int]
    MIN_ITEMS_FIELD_NUMBER: _ClassVar[int]
    MAX_ITEMS_FIELD_NUMBER: _ClassVar[int]
    uuid: bool
    non_empty: bool
    source_identity: bool
    timestamp: bool
    pattern: str
    minimum: int
    maximum: int
    min_items: int
    max_items: int
    def __init__(self, uuid: _Optional[bool] = ..., non_empty: _Optional[bool] = ..., source_identity: _Optional[bool] = ..., timestamp: _Optional[bool] = ..., pattern: _Optional[str] = ..., minimum: _Optional[int] = ..., maximum: _Optional[int] = ..., min_items: _Optional[int] = ..., max_items: _Optional[int] = ...) -> None: ...
