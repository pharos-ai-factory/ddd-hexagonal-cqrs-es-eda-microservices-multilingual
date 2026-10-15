"""Narrow schema-defined scalar values before exposing plain request contracts."""


def string_value(value: object) -> str:
    if not isinstance(value, str):
        raise ValueError("A string is required")
    return value


def integer_value(value: object, minimum: int, maximum: int) -> int:
    # Protobuf JSON represents 64-bit integers as decimal strings.
    if isinstance(value, str) and value.lstrip("-").isdigit():
        value = int(value)
    if isinstance(value, bool) or not isinstance(value, int) or not minimum <= value <= maximum:
        raise ValueError("An integer within the schema range is required")
    return value


def boolean_value(value: object) -> bool:
    if not isinstance(value, bool):
        raise ValueError("A boolean is required")
    return value
