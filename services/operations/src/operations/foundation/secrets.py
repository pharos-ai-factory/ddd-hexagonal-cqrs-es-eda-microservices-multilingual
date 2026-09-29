"""Resolve runtime credentials, with redacted configuration failures."""
import os
from pathlib import Path


def secret(name: str) -> str:
    """Accept NAME or NAME_FILE exclusively, preserving spaces in actual secrets."""
    if name in os.environ and name + "_FILE" in os.environ:
        raise RuntimeError(f"{name} and {name}_FILE are mutually exclusive")
    value = os.environ.get(name, "")
    if name + "_FILE" in os.environ:
        try:
            value = Path(os.environ[name + "_FILE"]).read_text().removesuffix("\n").removesuffix("\r")
        except (OSError, UnicodeError):
            raise RuntimeError(f"{name}_FILE cannot be read") from None
    if not value.strip():
        raise RuntimeError(f"{name} is required and must not be empty")
    return value
