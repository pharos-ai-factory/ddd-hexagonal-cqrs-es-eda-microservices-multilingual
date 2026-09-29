import base64
import re
from typing import TypedDict
from starlette.datastructures import QueryParams
from operations.foundation.application import Loaded
from operations.foundation.domain import Rejection, identifier
from operations.foundation.pagination import Page, PageRequest


def page_request(query: QueryParams, resource: str) -> PageRequest | None:
    if "limit" not in query and "cursor" not in query:
        return None
    invalid = Rejection("invalid_pagination", "Supply limit 1–100 and a cursor from this resource")
    raw = query.get("limit", "")
    if len(query.getlist("limit")) != 1 or len(query.getlist("cursor")) > 1 or not re.fullmatch(r"[1-9][0-9]{0,2}", raw):
        raise invalid
    limit = int(raw)
    if limit > 100:
        raise invalid
    after = None
    if "cursor" in query:
        cursor = query["cursor"]
        try:
            if len(cursor) > 1024:
                raise invalid
            decoded = base64.b64decode(cursor + "=" * (-len(cursor) % 4), altchars=b"-_", validate=True)
            if base64.urlsafe_b64encode(decoded).decode().rstrip("=") != cursor:
                raise invalid
            prefix = "1|" + resource + "|"
            text = decoded.decode()
            if not text.startswith(prefix):
                raise invalid
            after = identifier(text[len(prefix):])
        except (ValueError, Rejection) as error:
            raise invalid from error
    return PageRequest(limit, after)


class PageResponse[S](TypedDict):
    items: list[Loaded[S]]
    nextCursor: str | None


def page_response[S](page: Page[S], resource: str) -> PageResponse[S]:
    cursor = base64.urlsafe_b64encode(("1|" + resource + "|" + page.next_id).encode()).decode().rstrip("=") if page.next_id else None
    return {"items": page.items, "nextCursor": cursor}
