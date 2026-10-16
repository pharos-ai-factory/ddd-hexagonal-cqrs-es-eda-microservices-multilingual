"""Validate API connection reply addresses within the receiving context."""

import re

INSTANCE = re.compile(r"[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}")


def reply_route(owner: str, address: str | None) -> str:
    legacy = "reply." + owner
    if not address or address == legacy:
        return legacy
    if not address.startswith(legacy + ".") or not INSTANCE.fullmatch(address[len(legacy) + 1 :]):
        raise ValueError("Invalid owner reply address")
    return address
