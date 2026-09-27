from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Publication(_message.Message):
    __slots__ = ("event_id", "contract_version", "context", "aggregate_kind", "aggregate_id", "revision", "drink", "edition", "order", "ticket", "pickup", "account", "reward", "notification")
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    CONTRACT_VERSION_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_KIND_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_ID_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    DRINK_FIELD_NUMBER: _ClassVar[int]
    EDITION_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    TICKET_FIELD_NUMBER: _ClassVar[int]
    PICKUP_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    REWARD_FIELD_NUMBER: _ClassVar[int]
    NOTIFICATION_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    contract_version: int
    context: str
    aggregate_kind: str
    aggregate_id: str
    revision: int
    drink: Drink
    edition: Edition
    order: Order
    ticket: Ticket
    pickup: Pickup
    account: Account
    reward: Reward
    notification: Notification
    def __init__(self, event_id: _Optional[str] = ..., contract_version: _Optional[int] = ..., context: _Optional[str] = ..., aggregate_kind: _Optional[str] = ..., aggregate_id: _Optional[str] = ..., revision: _Optional[int] = ..., drink: _Optional[_Union[Drink, _Mapping]] = ..., edition: _Optional[_Union[Edition, _Mapping]] = ..., order: _Optional[_Union[Order, _Mapping]] = ..., ticket: _Optional[_Union[Ticket, _Mapping]] = ..., pickup: _Optional[_Union[Pickup, _Mapping]] = ..., account: _Optional[_Union[Account, _Mapping]] = ..., reward: _Optional[_Union[Reward, _Mapping]] = ..., notification: _Optional[_Union[Notification, _Mapping]] = ...) -> None: ...

class Drink(_message.Message):
    __slots__ = ("id", "name", "revision", "published")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    PUBLISHED_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    revision: int
    published: bool
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., revision: _Optional[int] = ..., published: _Optional[bool] = ...) -> None: ...

class Offer(_message.Message):
    __slots__ = ("code", "drink_id", "drink_revision", "name", "minor", "currency")
    CODE_FIELD_NUMBER: _ClassVar[int]
    DRINK_ID_FIELD_NUMBER: _ClassVar[int]
    DRINK_REVISION_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    MINOR_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    code: str
    drink_id: str
    drink_revision: int
    name: str
    minor: int
    currency: str
    def __init__(self, code: _Optional[str] = ..., drink_id: _Optional[str] = ..., drink_revision: _Optional[int] = ..., name: _Optional[str] = ..., minor: _Optional[int] = ..., currency: _Optional[str] = ...) -> None: ...

class Edition(_message.Message):
    __slots__ = ("id", "currency", "status", "offers")
    ID_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    OFFERS_FIELD_NUMBER: _ClassVar[int]
    id: str
    currency: str
    status: str
    offers: _containers.RepeatedCompositeFieldContainer[Offer]
    def __init__(self, id: _Optional[str] = ..., currency: _Optional[str] = ..., status: _Optional[str] = ..., offers: _Optional[_Iterable[_Union[Offer, _Mapping]]] = ...) -> None: ...

class Selection(_message.Message):
    __slots__ = ("offer_code", "name", "minor")
    OFFER_CODE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    MINOR_FIELD_NUMBER: _ClassVar[int]
    offer_code: str
    name: str
    minor: int
    def __init__(self, offer_code: _Optional[str] = ..., name: _Optional[str] = ..., minor: _Optional[int] = ...) -> None: ...

class Line(_message.Message):
    __slots__ = ("id", "selection", "quantity")
    ID_FIELD_NUMBER: _ClassVar[int]
    SELECTION_FIELD_NUMBER: _ClassVar[int]
    QUANTITY_FIELD_NUMBER: _ClassVar[int]
    id: str
    selection: Selection
    quantity: int
    def __init__(self, id: _Optional[str] = ..., selection: _Optional[_Union[Selection, _Mapping]] = ..., quantity: _Optional[int] = ...) -> None: ...

class Order(_message.Message):
    __slots__ = ("id", "customer_id", "edition_id", "currency", "status", "lines")
    ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    EDITION_ID_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    LINES_FIELD_NUMBER: _ClassVar[int]
    id: str
    customer_id: str
    edition_id: str
    currency: str
    status: str
    lines: _containers.RepeatedCompositeFieldContainer[Line]
    def __init__(self, id: _Optional[str] = ..., customer_id: _Optional[str] = ..., edition_id: _Optional[str] = ..., currency: _Optional[str] = ..., status: _Optional[str] = ..., lines: _Optional[_Iterable[_Union[Line, _Mapping]]] = ...) -> None: ...

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

class Pickup(_message.Message):
    __slots__ = ("id", "order_id", "customer_id", "code", "status")
    ID_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    id: str
    order_id: str
    customer_id: str
    code: str
    status: str
    def __init__(self, id: _Optional[str] = ..., order_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., code: _Optional[str] = ..., status: _Optional[str] = ...) -> None: ...

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
