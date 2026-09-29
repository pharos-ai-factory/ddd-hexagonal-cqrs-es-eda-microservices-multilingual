import pytest
from starlette.datastructures import QueryParams
from operations.adaptors.pagination import page_request, page_response
from operations.foundation.domain import Rejection
from operations.foundation.pagination import Page, PageRequest

ID = "00000000-0000-4000-8000-000000000001"


def test_pagination_is_optional_and_cursors_belong_to_their_resource() -> None:
    assert page_request(QueryParams(), "/tickets") is None
    page: Page[object] = Page([], ID)
    cursor = page_response(page, "/tickets")["nextCursor"]
    assert cursor
    assert page_request(QueryParams({"limit": "100", "cursor": cursor}), "/tickets") == PageRequest(100, ID)
    with pytest.raises(Rejection):
        page_request(QueryParams({"limit": "100", "cursor": cursor}), "/pickups")
    assert page_response(Page[object]([]), "/tickets")["nextCursor"] is None


@pytest.mark.parametrize("query", ["limit=0", "limit=101", "limit=-1", "limit=01", "limit=1.2",
    "limit=1&limit=2", "cursor=abc", "limit=1&cursor=", "limit=1&cursor=***", "limit=1&cursor=a&cursor=b"])
def test_invalid_page_parameters(query: str) -> None:
    with pytest.raises(Rejection):
        page_request(QueryParams(query), "/tickets")
