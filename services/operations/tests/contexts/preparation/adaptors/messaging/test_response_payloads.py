import pytest
from operations.contracts.requests.generated.preparation import ListTicketsQuery
from operations.contracts.requests.generated.preparation_responses import LoadedTicket, Tickets
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_requests_pb2 import Reply
from operations.contexts.preparation.adaptors.messaging.generated.response_payloads import (
    encode_list_tickets_response,
)
from operations.adaptors.messaging.generated.response_payloads import to_outcome


def test_generated_response_preserves_presence_and_nested_rows() -> None:
    row = LoadedTicket(
        exists=True,
        version=7,
        state={
            "id": "ticket",
            "orderId": "order",
            "customerId": "customer",
            "instructions": "Coffee",
            "status": "queued",
        },
    )
    values = Tickets(items=[row], paged=True, nextId="")
    result = encode_list_tickets_response(ListTicketsQuery(), {"tickets": values})
    restored = Reply.FromString(result.SerializeToString())
    assert restored.tickets.HasField("next_id") and restored.tickets.next_id == ""
    assert restored.tickets.items[0].version == 7
    assert restored.tickets.items[0].state.instructions == "Coffee"
    del values["nextId"]
    assert not encode_list_tickets_response(ListTicketsQuery(), {"tickets": values}).tickets.HasField(
        "next_id"
    )


def test_generated_outcome_preserves_rejection_and_rejects_unsafe_revision() -> None:
    result = to_outcome(
        {
            "aggregateId": "order",
            "version": 0,
            "status": "",
            "rejection": {"code": "not_found", "message": "Missing root"},
        }
    )
    assert result.rejection.code == "not_found" and result.version == 0
    with pytest.raises(ValueError):
        to_outcome({"aggregateId": "order", "version": 2**53, "status": "placed"})
