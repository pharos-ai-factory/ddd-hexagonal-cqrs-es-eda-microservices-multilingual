"""Provider-independent tactical primitives."""
import re
from typing import TypedDict


class RejectionDetail(TypedDict):
    code: str
    message: str


class Rejection(Exception):
    def __init__(self, code: str, message: str) -> None:
        super().__init__(message)
        self.code = code
        self.message = message

    def outcome(self) -> RejectionDetail:
        return {"code": self.code, "message": self.message}


class DomainError(Rejection):
    """Expected failure of a context-owned business rule."""


class CorruptState(ValueError):
    """Authoritative state cannot be restored; never record a business outcome."""


def identifier(value: object) -> str:
    if not isinstance(value, str) or not re.fullmatch(
        r"[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", value
    ):
        raise Rejection("invalid_id", "A canonical UUID is required")
    return value


def record(value: object) -> dict[str, object]:
    """Narrow an external object without trusting annotations on stored JSON."""
    if not isinstance(value, dict):
        raise ValueError("An object is required")
    result: dict[str, object] = {}
    for key, item in value.items():
        if not isinstance(key, str):
            raise ValueError("Object keys must be strings")
        result[key] = item
    return result


def text(value: object) -> str:
    if not isinstance(value, str):
        raise ValueError("A string is required")
    return value


def integer(value: object) -> int:
    if not isinstance(value, int) or isinstance(value, bool):
        raise ValueError("An integer is required")
    return value
