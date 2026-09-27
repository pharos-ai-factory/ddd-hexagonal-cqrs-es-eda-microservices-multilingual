"""Validate authoritative receipt shapes before returning a stored outcome."""
from operations.foundation.application import Outcome
from operations.foundation.domain import CorruptState, Rejection, RejectionDetail, identifier, integer, record, text


def outcome(value: object, target: str | None = None) -> Outcome:
    try:
        fields = record(value)
        version = integer(fields["version"])
        if version < 0:
            raise ValueError("Invalid recorded version")
        result = Outcome(aggregateId=identifier(fields["aggregateId"]), version=version, status=text(fields["status"]))
        if target is not None and result["aggregateId"] != target:
            raise ValueError("Recorded outcome identifies another aggregate")
        if "rejection" in fields:
            rejection = record(fields["rejection"])
            result["rejection"] = RejectionDetail(code=text(rejection["code"]), message=text(rejection["message"]))
            if not result["rejection"]["code"] or not result["rejection"]["message"]:
                raise ValueError("Incomplete recorded rejection")
        elif not result["status"]:
            raise ValueError("Missing recorded outcome status")
        return result
    except (Rejection, KeyError, TypeError, ValueError) as error:
        raise CorruptState("Corrupt recorded command outcome") from error
