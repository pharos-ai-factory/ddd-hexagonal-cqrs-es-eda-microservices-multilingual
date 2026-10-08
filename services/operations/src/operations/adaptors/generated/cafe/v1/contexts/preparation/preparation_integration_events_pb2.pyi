from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class DrinksReady(_message.Message):
    __slots__ = ("order_id", "customer_id")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    customer_id: str
    def __init__(self, order_id: _Optional[str] = ..., customer_id: _Optional[str] = ...) -> None: ...
