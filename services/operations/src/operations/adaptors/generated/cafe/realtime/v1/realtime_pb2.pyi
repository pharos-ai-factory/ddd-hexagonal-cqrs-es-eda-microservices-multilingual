from operations.adaptors.generated.cafe.realtime.v1.contexts.menu import menu_snapshots_pb2 as _menu_snapshots_pb2
from operations.adaptors.generated.cafe.realtime.v1.contexts.ordering import ordering_snapshots_pb2 as _ordering_snapshots_pb2
from operations.adaptors.generated.cafe.realtime.v1.contexts.preparation import preparation_snapshots_pb2 as _preparation_snapshots_pb2
from operations.adaptors.generated.cafe.realtime.v1.contexts.collection import collection_snapshots_pb2 as _collection_snapshots_pb2
from operations.adaptors.generated.cafe.realtime.v1.contexts.loyalty import loyalty_snapshots_pb2 as _loyalty_snapshots_pb2
from operations.adaptors.generated.cafe.realtime.v1.contexts.communication import communication_snapshots_pb2 as _communication_snapshots_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
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
    drink: _menu_snapshots_pb2.Drink
    edition: _menu_snapshots_pb2.Edition
    order: _ordering_snapshots_pb2.Order
    ticket: _preparation_snapshots_pb2.Ticket
    pickup: _collection_snapshots_pb2.Pickup
    account: _loyalty_snapshots_pb2.Account
    reward: _loyalty_snapshots_pb2.Reward
    notification: _communication_snapshots_pb2.Notification
    def __init__(self, event_id: _Optional[str] = ..., contract_version: _Optional[int] = ..., context: _Optional[str] = ..., aggregate_kind: _Optional[str] = ..., aggregate_id: _Optional[str] = ..., revision: _Optional[int] = ..., drink: _Optional[_Union[_menu_snapshots_pb2.Drink, _Mapping]] = ..., edition: _Optional[_Union[_menu_snapshots_pb2.Edition, _Mapping]] = ..., order: _Optional[_Union[_ordering_snapshots_pb2.Order, _Mapping]] = ..., ticket: _Optional[_Union[_preparation_snapshots_pb2.Ticket, _Mapping]] = ..., pickup: _Optional[_Union[_collection_snapshots_pb2.Pickup, _Mapping]] = ..., account: _Optional[_Union[_loyalty_snapshots_pb2.Account, _Mapping]] = ..., reward: _Optional[_Union[_loyalty_snapshots_pb2.Reward, _Mapping]] = ..., notification: _Optional[_Union[_communication_snapshots_pb2.Notification, _Mapping]] = ...) -> None: ...
