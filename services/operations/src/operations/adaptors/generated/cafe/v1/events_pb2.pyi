from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Event(_message.Message):
    __slots__ = ("id", "name", "context", "visibility", "contract_version", "aggregate_kind", "aggregate_id", "aggregate_version", "correlation_id", "causation_id", "occurred_at", "payload")
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
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
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
    payload: bytes
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., context: _Optional[str] = ..., visibility: _Optional[str] = ..., contract_version: _Optional[int] = ..., aggregate_kind: _Optional[str] = ..., aggregate_id: _Optional[str] = ..., aggregate_version: _Optional[int] = ..., correlation_id: _Optional[str] = ..., causation_id: _Optional[str] = ..., occurred_at: _Optional[str] = ..., payload: _Optional[bytes] = ...) -> None: ...
