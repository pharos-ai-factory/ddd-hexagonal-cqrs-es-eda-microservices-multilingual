from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

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
