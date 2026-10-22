"""Application delivery targets survive the Python realtime codec."""

from operations.adaptors.generated.cafe.realtime.v1.realtime_pb2 import Publication
from operations.adaptors.realtime.message_codec import encode, encode_rejection
from operations.foundation.application import context_target, principal_target

ID = "00000000-0000-4000-8000-000000000001"


def test_context_and_principal_targets_are_encoded() -> None:
    _, body = encode("preparation", "ticket", ID, 1, context_target("preparation"), {"id": ID})
    assert Publication.FromString(body).target.context == "preparation"

    _, body = encode_rejection(
        "preparation",
        "ticket",
        ID,
        0,
        principal_target(ID),
        ID,
        ID,
        "not_found",
        "missing",
    )
    assert Publication.FromString(body).target.principal == ID
