from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CommandEnvelope(_message.Message):
    __slots__ = ("id", "consumer", "source_event_id", "source_hash", "target", "correlation_id", "receipt_material", "accept_order")
    ID_FIELD_NUMBER: _ClassVar[int]
    CONSUMER_FIELD_NUMBER: _ClassVar[int]
    SOURCE_EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    SOURCE_HASH_FIELD_NUMBER: _ClassVar[int]
    TARGET_FIELD_NUMBER: _ClassVar[int]
    CORRELATION_ID_FIELD_NUMBER: _ClassVar[int]
    RECEIPT_MATERIAL_FIELD_NUMBER: _ClassVar[int]
    ACCEPT_ORDER_FIELD_NUMBER: _ClassVar[int]
    id: str
    consumer: str
    source_event_id: str
    source_hash: str
    target: str
    correlation_id: str
    receipt_material: bytes
    accept_order: AcceptOrderCommand
    def __init__(self, id: _Optional[str] = ..., consumer: _Optional[str] = ..., source_event_id: _Optional[str] = ..., source_hash: _Optional[str] = ..., target: _Optional[str] = ..., correlation_id: _Optional[str] = ..., receipt_material: _Optional[bytes] = ..., accept_order: _Optional[_Union[AcceptOrderCommand, _Mapping]] = ...) -> None: ...

class AcceptOrderCommand(_message.Message):
    __slots__ = ("order_id", "customer_id", "instructions")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    INSTRUCTIONS_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    customer_id: str
    instructions: str
    def __init__(self, order_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., instructions: _Optional[str] = ...) -> None: ...
