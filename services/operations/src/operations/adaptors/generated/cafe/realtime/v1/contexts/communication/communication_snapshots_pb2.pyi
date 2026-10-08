from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Notification(_message.Message):
    __slots__ = ("id", "recipient", "subject", "body", "status", "provider_receipt")
    ID_FIELD_NUMBER: _ClassVar[int]
    RECIPIENT_FIELD_NUMBER: _ClassVar[int]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    PROVIDER_RECEIPT_FIELD_NUMBER: _ClassVar[int]
    id: str
    recipient: str
    subject: str
    body: str
    status: str
    provider_receipt: str
    def __init__(self, id: _Optional[str] = ..., recipient: _Optional[str] = ..., subject: _Optional[str] = ..., body: _Optional[str] = ..., status: _Optional[str] = ..., provider_receipt: _Optional[str] = ...) -> None: ...
