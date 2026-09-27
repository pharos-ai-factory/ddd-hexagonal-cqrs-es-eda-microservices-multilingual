from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Event(_message.Message):
    __slots__ = ("id", "name", "context", "visibility", "contract_version", "aggregate_kind", "aggregate_id", "aggregate_version", "correlation_id", "causation_id", "occurred_at", "drink_published", "menu_published", "order_placed", "drinks_ready", "pickup_opened", "order_collected", "reward_earned", "reward_issued", "notification_requested")
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
    DRINK_PUBLISHED_FIELD_NUMBER: _ClassVar[int]
    MENU_PUBLISHED_FIELD_NUMBER: _ClassVar[int]
    ORDER_PLACED_FIELD_NUMBER: _ClassVar[int]
    DRINKS_READY_FIELD_NUMBER: _ClassVar[int]
    PICKUP_OPENED_FIELD_NUMBER: _ClassVar[int]
    ORDER_COLLECTED_FIELD_NUMBER: _ClassVar[int]
    REWARD_EARNED_FIELD_NUMBER: _ClassVar[int]
    REWARD_ISSUED_FIELD_NUMBER: _ClassVar[int]
    NOTIFICATION_REQUESTED_FIELD_NUMBER: _ClassVar[int]
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
    drink_published: DrinkPublished
    menu_published: MenuPublished
    order_placed: OrderPlaced
    drinks_ready: DrinksReady
    pickup_opened: PickupOpened
    order_collected: OrderCollected
    reward_earned: RewardEarned
    reward_issued: RewardIssued
    notification_requested: NotificationRequested
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., context: _Optional[str] = ..., visibility: _Optional[str] = ..., contract_version: _Optional[int] = ..., aggregate_kind: _Optional[str] = ..., aggregate_id: _Optional[str] = ..., aggregate_version: _Optional[int] = ..., correlation_id: _Optional[str] = ..., causation_id: _Optional[str] = ..., occurred_at: _Optional[str] = ..., drink_published: _Optional[_Union[DrinkPublished, _Mapping]] = ..., menu_published: _Optional[_Union[MenuPublished, _Mapping]] = ..., order_placed: _Optional[_Union[OrderPlaced, _Mapping]] = ..., drinks_ready: _Optional[_Union[DrinksReady, _Mapping]] = ..., pickup_opened: _Optional[_Union[PickupOpened, _Mapping]] = ..., order_collected: _Optional[_Union[OrderCollected, _Mapping]] = ..., reward_earned: _Optional[_Union[RewardEarned, _Mapping]] = ..., reward_issued: _Optional[_Union[RewardIssued, _Mapping]] = ..., notification_requested: _Optional[_Union[NotificationRequested, _Mapping]] = ...) -> None: ...

class DrinkPublished(_message.Message):
    __slots__ = ("drink_id", "name", "revision")
    DRINK_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    drink_id: str
    name: str
    revision: int
    def __init__(self, drink_id: _Optional[str] = ..., name: _Optional[str] = ..., revision: _Optional[int] = ...) -> None: ...

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

class MenuPublished(_message.Message):
    __slots__ = ("edition_id", "currency", "offers")
    EDITION_ID_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    OFFERS_FIELD_NUMBER: _ClassVar[int]
    edition_id: str
    currency: str
    offers: _containers.RepeatedCompositeFieldContainer[Offer]
    def __init__(self, edition_id: _Optional[str] = ..., currency: _Optional[str] = ..., offers: _Optional[_Iterable[_Union[Offer, _Mapping]]] = ...) -> None: ...

class Line(_message.Message):
    __slots__ = ("id", "offer_code", "name", "quantity", "minor")
    ID_FIELD_NUMBER: _ClassVar[int]
    OFFER_CODE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    QUANTITY_FIELD_NUMBER: _ClassVar[int]
    MINOR_FIELD_NUMBER: _ClassVar[int]
    id: str
    offer_code: str
    name: str
    quantity: int
    minor: int
    def __init__(self, id: _Optional[str] = ..., offer_code: _Optional[str] = ..., name: _Optional[str] = ..., quantity: _Optional[int] = ..., minor: _Optional[int] = ...) -> None: ...

class OrderPlaced(_message.Message):
    __slots__ = ("order_id", "customer_id", "edition_id", "currency", "lines")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    EDITION_ID_FIELD_NUMBER: _ClassVar[int]
    CURRENCY_FIELD_NUMBER: _ClassVar[int]
    LINES_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    customer_id: str
    edition_id: str
    currency: str
    lines: _containers.RepeatedCompositeFieldContainer[Line]
    def __init__(self, order_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., edition_id: _Optional[str] = ..., currency: _Optional[str] = ..., lines: _Optional[_Iterable[_Union[Line, _Mapping]]] = ...) -> None: ...

class DrinksReady(_message.Message):
    __slots__ = ("order_id", "customer_id")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    customer_id: str
    def __init__(self, order_id: _Optional[str] = ..., customer_id: _Optional[str] = ...) -> None: ...

class PickupOpened(_message.Message):
    __slots__ = ("pickup_id", "order_id", "customer_id", "collection_code")
    PICKUP_ID_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_CODE_FIELD_NUMBER: _ClassVar[int]
    pickup_id: str
    order_id: str
    customer_id: str
    collection_code: str
    def __init__(self, pickup_id: _Optional[str] = ..., order_id: _Optional[str] = ..., customer_id: _Optional[str] = ..., collection_code: _Optional[str] = ...) -> None: ...

class OrderCollected(_message.Message):
    __slots__ = ("order_id", "customer_id")
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CUSTOMER_ID_FIELD_NUMBER: _ClassVar[int]
    order_id: str
    customer_id: str
    def __init__(self, order_id: _Optional[str] = ..., customer_id: _Optional[str] = ...) -> None: ...

class RewardEarned(_message.Message):
    __slots__ = ("account_id", "grant_id", "benefit", "valid_days")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    GRANT_ID_FIELD_NUMBER: _ClassVar[int]
    BENEFIT_FIELD_NUMBER: _ClassVar[int]
    VALID_DAYS_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    grant_id: str
    benefit: str
    valid_days: int
    def __init__(self, account_id: _Optional[str] = ..., grant_id: _Optional[str] = ..., benefit: _Optional[str] = ..., valid_days: _Optional[int] = ...) -> None: ...

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

class NotificationRequested(_message.Message):
    __slots__ = ("notification_id",)
    NOTIFICATION_ID_FIELD_NUMBER: _ClassVar[int]
    notification_id: str
    def __init__(self, notification_id: _Optional[str] = ...) -> None: ...
