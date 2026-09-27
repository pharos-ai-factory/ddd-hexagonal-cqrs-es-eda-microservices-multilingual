from hashlib import sha256
from uuid import UUID, uuid4


def new_id() -> str:
    return str(uuid4())


def derived_id(purpose: str, key: str) -> str:
    raw = bytearray(sha256(f"cafe-reference/v1\0{purpose}\0{key}".encode()).digest()[:16])
    raw[6] = (raw[6] & 15) | 128
    raw[8] = (raw[8] & 63) | 128
    return str(UUID(bytes=bytes(raw)))
