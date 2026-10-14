"""Keep event-specific literals out of shared handwritten transport code."""

import re

SHARED_CODECS = {
    "services/operations/src/operations/adaptors/messaging/event_codec.py",
    "services/operations/src/operations/adaptors/messaging/event_payload_codec.py",
    "services/engagement/src/adaptors/messaging/event-codec.ts",
    "services/engagement/src/adaptors/messaging/event-payload-codec.ts",
    "services/engagement/src/adaptors/messaging/event-envelope.ts",
    "services/storefront/contracts/events/protobuf/codec.go",
    "services/storefront/contracts/events/protobuf/validation.go",
}


def violation(path, source, published_names):
    if path in SHARED_CODECS and set(re.findall(r"""["']([^"'\n]+)["']""", source)) & set(published_names):
        return "event-specific mapping in shared transport; use schema generation or a context adaptor"
    return None
