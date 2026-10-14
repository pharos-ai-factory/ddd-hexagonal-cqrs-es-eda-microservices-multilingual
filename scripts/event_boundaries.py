"""Keep event-specific literals out of shared handwritten transport code."""

import re

SHARED_CODECS = {
    "services/operations/src/operations/adaptors/codec.py",
    "services/operations/src/operations/adaptors/event_payload.py",
    "services/engagement/src/adaptors/codec.ts",
    "services/engagement/src/adaptors/event-payload.ts",
    "services/engagement/src/adaptors/event-envelope.ts",
    "services/storefront/contracts/events/protobuf/codec.go",
    "services/storefront/contracts/events/protobuf/validation.go",
}


def violation(path, source, published_names):
    if path in SHARED_CODECS and set(re.findall(r"""["']([^"'\n]+)["']""", source)) & set(published_names):
        return "event-specific mapping in shared transport; use schema generation or a context adaptor"
    return None
