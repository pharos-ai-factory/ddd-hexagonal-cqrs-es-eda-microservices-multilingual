from operations.adaptors.generated.cafe.realtime.v1.contexts.preparation import preparation_snapshots_pb2 as _preparation_snapshots_pb2
from operations.adaptors.generated.cafe.realtime.v1.contexts.collection import collection_snapshots_pb2 as _collection_snapshots_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Publication(_message.Message):
    __slots__ = ("event_id", "contract_version", "context", "aggregate_kind", "aggregate_id", "revision", "ticket", "pickup")
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    CONTRACT_VERSION_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_KIND_FIELD_NUMBER: _ClassVar[int]
    AGGREGATE_ID_FIELD_NUMBER: _ClassVar[int]
    REVISION_FIELD_NUMBER: _ClassVar[int]
    TICKET_FIELD_NUMBER: _ClassVar[int]
    PICKUP_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    contract_version: int
    context: str
    aggregate_kind: str
    aggregate_id: str
    revision: int
    ticket: _preparation_snapshots_pb2.Ticket
    pickup: _collection_snapshots_pb2.Pickup
    def __init__(self, event_id: _Optional[str] = ..., contract_version: _Optional[int] = ..., context: _Optional[str] = ..., aggregate_kind: _Optional[str] = ..., aggregate_id: _Optional[str] = ..., revision: _Optional[int] = ..., ticket: _Optional[_Union[_preparation_snapshots_pb2.Ticket, _Mapping]] = ..., pickup: _Optional[_Union[_collection_snapshots_pb2.Pickup, _Mapping]] = ...) -> None: ...
