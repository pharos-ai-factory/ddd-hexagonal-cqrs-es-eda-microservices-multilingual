from google.protobuf import descriptor_pb2 as _descriptor_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Visibility(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    VISIBILITY_UNSPECIFIED: _ClassVar[Visibility]
    PUBLIC: _ClassVar[Visibility]
    PRIVATE: _ClassVar[Visibility]
VISIBILITY_UNSPECIFIED: Visibility
PUBLIC: Visibility
PRIVATE: Visibility
REQUIRED_INPUT_FIELD_NUMBER: _ClassVar[int]
required_input: _descriptor.FieldDescriptor
CONTRACT_FIELD_NUMBER: _ClassVar[int]
contract: _descriptor.FieldDescriptor

class RequestContract(_message.Message):
    __slots__ = ("owner", "visibility")
    OWNER_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    owner: str
    visibility: Visibility
    def __init__(self, owner: _Optional[str] = ..., visibility: _Optional[_Union[Visibility, str]] = ...) -> None: ...
