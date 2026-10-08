from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PageRequest(_message.Message):
    __slots__ = ("limit", "after")
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    AFTER_FIELD_NUMBER: _ClassVar[int]
    limit: int
    after: str
    def __init__(self, limit: _Optional[int] = ..., after: _Optional[str] = ...) -> None: ...

class CommandMetadata(_message.Message):
    __slots__ = ("command_id", "aggregate_id", "expected_version", "correlation_id")
    COMMAND_ID_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_ID_FIELD_NUMBER: _ClassVar[int]
    EXPECTED_VERSION_FIELD_NUMBER: _ClassVar[int]
    CORRELATION_ID_FIELD_NUMBER: _ClassVar[int]
    command_id: str
    aggregate_id: str
    expected_version: int
    correlation_id: str
    def __init__(self, command_id: _Optional[str] = ..., aggregate_id: _Optional[str] = ..., expected_version: _Optional[int] = ..., correlation_id: _Optional[str] = ...) -> None: ...

class Rejection(_message.Message):
    __slots__ = ("code", "message")
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    code: str
    message: str
    def __init__(self, code: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...

class Outcome(_message.Message):
    __slots__ = ("aggregate_id", "version", "status", "rejection")
    AGGREGATE_ID_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    REJECTION_FIELD_NUMBER: _ClassVar[int]
    aggregate_id: str
    version: int
    status: str
    rejection: Rejection
    def __init__(self, aggregate_id: _Optional[str] = ..., version: _Optional[int] = ..., status: _Optional[str] = ..., rejection: _Optional[_Union[Rejection, _Mapping]] = ...) -> None: ...

class RequestError(_message.Message):
    __slots__ = ("code", "message")
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    code: str
    message: str
    def __init__(self, code: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...
