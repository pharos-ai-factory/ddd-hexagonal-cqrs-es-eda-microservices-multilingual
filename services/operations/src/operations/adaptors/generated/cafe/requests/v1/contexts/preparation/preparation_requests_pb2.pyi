from operations.adaptors.generated.cafe.requests.v1 import common_pb2 as _common_pb2
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation import preparation_queries_pb2 as _preparation_queries_pb2
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation import preparation_replies_pb2 as _preparation_replies_pb2
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation import preparation_commands_pb2 as _preparation_commands_pb2
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
    __slots__ = ("metadata", "start_preparation", "complete_preparation")
    METADATA_FIELD_NUMBER: _ClassVar[int]
    START_PREPARATION_FIELD_NUMBER: _ClassVar[int]
    COMPLETE_PREPARATION_FIELD_NUMBER: _ClassVar[int]
    metadata: _common_pb2.CommandMetadata
    start_preparation: _preparation_commands_pb2.StartPreparation
    complete_preparation: _preparation_commands_pb2.CompletePreparation
    def __init__(self, metadata: _Optional[_Union[_common_pb2.CommandMetadata, _Mapping]] = ..., start_preparation: _Optional[_Union[_preparation_commands_pb2.StartPreparation, _Mapping]] = ..., complete_preparation: _Optional[_Union[_preparation_commands_pb2.CompletePreparation, _Mapping]] = ...) -> None: ...

class Query(_message.Message):
    __slots__ = ("list_tickets", "get_ticket")
    LIST_TICKETS_FIELD_NUMBER: _ClassVar[int]
    GET_TICKET_FIELD_NUMBER: _ClassVar[int]
    list_tickets: _preparation_queries_pb2.ListTickets
    get_ticket: _preparation_queries_pb2.GetTicket
    def __init__(self, list_tickets: _Optional[_Union[_preparation_queries_pb2.ListTickets, _Mapping]] = ..., get_ticket: _Optional[_Union[_preparation_queries_pb2.GetTicket, _Mapping]] = ...) -> None: ...

class Reply(_message.Message):
    __slots__ = ("contract_version", "request_id", "context", "outcome", "error", "ticket", "tickets")
    CONTRACT_VERSION_FIELD_NUMBER: _ClassVar[int]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    CONTEXT_FIELD_NUMBER: _ClassVar[int]
    OUTCOME_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    TICKET_FIELD_NUMBER: _ClassVar[int]
    TICKETS_FIELD_NUMBER: _ClassVar[int]
    contract_version: int
    request_id: str
    context: str
    outcome: _common_pb2.Outcome
    error: _common_pb2.RequestError
    ticket: _preparation_replies_pb2.LoadedTicket
    tickets: _preparation_replies_pb2.Tickets
    def __init__(self, contract_version: _Optional[int] = ..., request_id: _Optional[str] = ..., context: _Optional[str] = ..., outcome: _Optional[_Union[_common_pb2.Outcome, _Mapping]] = ..., error: _Optional[_Union[_common_pb2.RequestError, _Mapping]] = ..., ticket: _Optional[_Union[_preparation_replies_pb2.LoadedTicket, _Mapping]] = ..., tickets: _Optional[_Union[_preparation_replies_pb2.Tickets, _Mapping]] = ...) -> None: ...
