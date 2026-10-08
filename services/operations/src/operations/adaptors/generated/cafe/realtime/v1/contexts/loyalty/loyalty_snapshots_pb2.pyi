from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Grant(_message.Message):
    __slots__ = ("id", "account_id", "benefit", "valid_days")
    ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    BENEFIT_FIELD_NUMBER: _ClassVar[int]
    VALID_DAYS_FIELD_NUMBER: _ClassVar[int]
    id: str
    account_id: str
    benefit: str
    valid_days: int
    def __init__(self, id: _Optional[str] = ..., account_id: _Optional[str] = ..., benefit: _Optional[str] = ..., valid_days: _Optional[int] = ...) -> None: ...

class Account(_message.Message):
    __slots__ = ("id", "stamp_balance", "collections", "grants_earned", "last_grant")
    ID_FIELD_NUMBER: _ClassVar[int]
    STAMP_BALANCE_FIELD_NUMBER: _ClassVar[int]
    COLLECTIONS_FIELD_NUMBER: _ClassVar[int]
    GRANTS_EARNED_FIELD_NUMBER: _ClassVar[int]
    LAST_GRANT_FIELD_NUMBER: _ClassVar[int]
    id: str
    stamp_balance: int
    collections: int
    grants_earned: int
    last_grant: Grant
    def __init__(self, id: _Optional[str] = ..., stamp_balance: _Optional[int] = ..., collections: _Optional[int] = ..., grants_earned: _Optional[int] = ..., last_grant: _Optional[_Union[Grant, _Mapping]] = ...) -> None: ...

class Reward(_message.Message):
    __slots__ = ("id", "grant_id", "customer_id", "benefit", "status", "expires_at", "redeemed_for")
    ID_FIELD_NUMBER: _ClassVar[int]
    GRANT_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    BENEFIT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    REDEEMED_FOR_FIELD_NUMBER: _ClassVar[int]
    id: str
    grant_id: str
    customer_id: str
    benefit: str
    status: str
    expires_at: str
    redeemed_for: str
    def __init__(self, id: _Optional[str] = ..., grant_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., benefit: _Optional[str] = ..., status: _Optional[str] = ..., expires_at: _Optional[str] = ..., redeemed_for: _Optional[str] = ...) -> None: ...
