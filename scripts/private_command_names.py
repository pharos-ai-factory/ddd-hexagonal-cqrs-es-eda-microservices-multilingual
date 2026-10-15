"""Read private command identifiers from their annotated owner Protobuf envelopes."""

from request_payloads import load


def read(owner, root):
    return [
        (m["field"], m["route"])
        for m in load(root)["payloads"]
        if m["owner"] == owner and m["visibility"] == "PRIVATE"
    ]
