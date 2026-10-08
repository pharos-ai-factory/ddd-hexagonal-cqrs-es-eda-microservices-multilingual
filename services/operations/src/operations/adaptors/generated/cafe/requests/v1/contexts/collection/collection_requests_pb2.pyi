from operations.adaptors.generated.cafe.requests.v1 import common_pb2 as _common_pb2
from operations.adaptors.generated.cafe.requests.v1.contexts.collection import collection_queries_pb2 as _collection_queries_pb2
from operations.adaptors.generated.cafe.requests.v1.contexts.collection import collection_replies_pb2 as _collection_replies_pb2
from operations.adaptors.generated.cafe.requests.v1.contexts.collection import collection_commands_pb2 as _collection_commands_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Request(_message.Message):
    __slots__ = ("contract_version", "request_id", "context", "command", "query")
    CONTRACT_VERSION_FIELD_NUMBER: _ClassVar[int]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    COMMAND_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    contract_version: int
    request_id: str
    context: str
    command: Command
    query: Query
    def __init__(self, contract_version: _Optional[int] = ..., request_id: _Optional[str] = ..., context: _Optional[str] = ..., command: _Optional[_Union[Command, _Mapping]] = ..., query: _Optional[_Union[Query, _Mapping]] = ...) -> None: ...

class Command(_message.Message):
    __slots__ = ("metadata", "collect_order")
    METADATA_FIELD_NUMBER: _ClassVar[int]
    COLLECT_ORDER_FIELD_NUMBER: _ClassVar[int]
    metadata: _common_pb2.CommandMetadata
    collect_order: _collection_commands_pb2.CollectOrder
    def __init__(self, metadata: _Optional[_Union[_common_pb2.CommandMetadata, _Mapping]] = ..., collect_order: _Optional[_Union[_collection_commands_pb2.CollectOrder, _Mapping]] = ...) -> None: ...

class Query(_message.Message):
    __slots__ = ("list_pickups", "get_pickup")
    LIST_PICKUPS_FIELD_NUMBER: _ClassVar[int]
    GET_PICKUP_FIELD_NUMBER: _ClassVar[int]
    list_pickups: _collection_queries_pb2.ListPickups
    get_pickup: _collection_queries_pb2.GetPickup
    def __init__(self, list_pickups: _Optional[_Union[_collection_queries_pb2.ListPickups, _Mapping]] = ..., get_pickup: _Optional[_Union[_collection_queries_pb2.GetPickup, _Mapping]] = ...) -> None: ...

class Reply(_message.Message):
    __slots__ = ("contract_version", "request_id", "context", "outcome", "error", "pickup", "pickups")
    CONTRACT_VERSION_FIELD_NUMBER: _ClassVar[int]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    OUTCOME_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    PICKUP_FIELD_NUMBER: _ClassVar[int]
    PICKUPS_FIELD_NUMBER: _ClassVar[int]
    contract_version: int
    request_id: str
    context: str
    outcome: _common_pb2.Outcome
    error: _common_pb2.RequestError
    pickup: _collection_replies_pb2.LoadedPickup
    pickups: _collection_replies_pb2.Pickups
    def __init__(self, contract_version: _Optional[int] = ..., request_id: _Optional[str] = ..., context: _Optional[str] = ..., outcome: _Optional[_Union[_common_pb2.Outcome, _Mapping]] = ..., error: _Optional[_Union[_common_pb2.RequestError, _Mapping]] = ..., pickup: _Optional[_Union[_collection_replies_pb2.LoadedPickup, _Mapping]] = ..., pickups: _Optional[_Union[_collection_replies_pb2.Pickups, _Mapping]] = ...) -> None: ...
