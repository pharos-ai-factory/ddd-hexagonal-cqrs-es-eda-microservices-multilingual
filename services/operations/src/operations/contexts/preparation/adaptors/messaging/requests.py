"""Explicit mapping between preparation wire contracts and application types."""
from google.protobuf.message import Message
from operations.foundation.application import Loaded
from operations.contexts.preparation.domain import TicketSnapshot
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_replies_pb2 import Ticket, LoadedTicket, Tickets
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_requests_pb2 import Reply
from operations.contexts.preparation.application.commands.start_preparation import StartPreparationCommand
from operations.contexts.preparation.application.commands.complete_preparation import CompletePreparationCommand
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_commands_pb2 import StartPreparation, CompletePreparation


def start(value: Message) -> StartPreparationCommand:
    if not isinstance(value, StartPreparation):
        raise ValueError("Unexpected start preparation payload")
    return StartPreparationCommand()


def complete(value: Message) -> CompletePreparationCommand:
    if not isinstance(value, CompletePreparation):
        raise ValueError("Unexpected complete preparation payload")
    return CompletePreparationCommand()


def item(loaded: Loaded[TicketSnapshot]) -> LoadedTicket:
    s = loaded["state"]
    return LoadedTicket(exists=loaded["exists"], version=loaded["version"], state=Ticket(id=s["id"], order_id=s["orderId"], customer_id=s["customerId"], instructions=s["instructions"], status=s["status"]))


def listing(items: list[Loaded[TicketSnapshot]], paged: bool, next_id: str | None) -> Reply:
    result = Tickets(items=[item(value) for value in items], paged=paged)
    if next_id is not None:
        result.next_id = next_id
    return Reply(tickets=result)
