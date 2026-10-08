from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class RewardIssued(_message.Message):
    __slots__ = ("reward_id", "customer_id", "benefit", "expires_at")
    REWARD_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    BENEFIT_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    reward_id: str
    customer_id: str
    benefit: str
    expires_at: str
    def __init__(self, reward_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., benefit: _Optional[str] = ..., expires_at: _Optional[str] = ...) -> None: ...
