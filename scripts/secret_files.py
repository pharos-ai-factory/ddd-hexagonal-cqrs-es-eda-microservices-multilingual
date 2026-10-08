"""Protected development configuration and per-developer secret-file references."""
import os
from pathlib import Path
import stat
import tempfile
from urllib.parse import quote


def atomic_private_write(path, text):
    """Replace a regular file atomically; existing permissive modes become 0600."""
    path = Path(path)
    if path.is_symlink() or (path.exists() and not path.is_file()):
        raise ValueError("Configuration target must be a regular file")
    path.parent.mkdir(parents=True, exist_ok=True)
    descriptor, temporary = tempfile.mkstemp(prefix=".secret-", dir=path.parent)
    try:
        with os.fdopen(descriptor, "w") as stream:
            os.fchmod(stream.fileno(), 0o600)
            stream.write(text)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def read_configuration(path):
    if path.is_symlink():
        raise ValueError("Configuration target must be a regular file")
    return dict(line.split("=", 1) for line in path.read_text().splitlines()
                if line and not line.startswith("#"))


def resolve_files(values, directory):
    """Keep file references on disk; expand only in the launcher process."""
    result = dict(values)
    for key, filename in values.items():
        if not key.endswith("_FILE"):
            continue
        name = key.removesuffix("_FILE")
        if name in values:
            raise ValueError(f"{name} and {key} are mutually exclusive")
        try:
            path = Path(filename)
            if not path.is_absolute():
                path = directory / path
            info = path.lstat()
            if not stat.S_ISREG(info.st_mode) or info.st_mode & 0o077:
                raise OSError()
            value = path.read_text().removesuffix("\n").removesuffix("\r")
            if not value.strip():
                raise OSError()
        except (OSError, UnicodeError):
            raise ValueError(f"{key} requires a readable, non-empty private regular file") from None
        result.pop(key)
        result[name] = value
    return result


def compose_environment(values, owners):
    result = dict(os.environ, **values)
    for owner in owners:
        prefix = owner.upper()
        password = quote(values[prefix + "_DB_PASSWORD"], safe="")
        result[prefix + "_DATABASE_URL"] = (
            f"postgres://cafe_{owner}:{password}@postgres:5432/cafe_{owner}?sslmode=disable")
        password = quote(values[prefix + "_BROKER_PASSWORD"], safe="")
        result[prefix + "_BROKER_URL"] = f"amqp://cafe_{owner}:{password}@rabbitmq:5672/reference"
    password = quote(values["API_BROKER_PASSWORD"], safe="")
    result["API_BROKER_URL"] = f"amqp://cafe_api:{password}@rabbitmq:5672/reference"
    password = quote(values["BROKER_PASSWORD"], safe="")
    result["BROKER_ADMIN_URL"] = f"amqp://administrator:{password}@rabbitmq:5672/reference"
    return result
