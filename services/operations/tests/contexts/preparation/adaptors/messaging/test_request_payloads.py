"""Generated query inputs preserve wire presence before owner semantic checks."""

import pytest
from operations.contexts.preparation.adaptors.messaging.generated.request_payloads import list_tickets
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_queries_pb2 import (
    ListTickets,
)
from operations.adaptors.generated.cafe.requests.v1.common_pb2 import PageRequest
from operations.contracts.requests.generated.shared import PageRequest as PlainPageRequest
from operations.adaptors.messaging.request_values import integer_value


def test_complete_list_and_present_zero_limit_have_distinct_inputs() -> None:
    assert list_tickets(ListTickets()).page is None
    assert list_tickets(ListTickets(page=PageRequest(limit=0))).page == PlainPageRequest(0)
    assert list_tickets(ListTickets(page=PageRequest(limit=2, after=""))).page == PlainPageRequest(2, "")


def test_integer_wire_range_and_explicit_zero() -> None:
    assert integer_value(0, 0, 100) == 0
    assert integer_value("100", 0, 100) == 100
    for value in (True, 1.5, -1, 101):
        with pytest.raises(ValueError):
            integer_value(value, 0, 100)
