"""Technical backlog evidence and redacted, process-lifetime worker failures."""
from collections.abc import Mapping
from datetime import datetime, timezone
import json
import logging
from threading import Lock
from typing import TypedDict
from operations.adaptors.postgres import PostgresContextDatabase


class Failure(TypedDict):
    """Stores redacted delivery failure evidence without exception messages or credentials."""
    at: str
    error_class: str
    event: str
    correlation: str


class Worker(TypedDict):
    """Reports process-local delivery failures and confirmed dead-letter transfers."""
    owner: str
    name: str
    failures: int
    deadLetterTransfers: int
    lastFailure: Failure


_lock = Lock()
_workers: dict[str, Worker] = {}


def failure(owner: str, name: str, error: Exception, event: str = "", correlation: str = "", dead: bool = False) -> None:
    last = Failure(at=datetime.now(timezone.utc).isoformat(), error_class=type(error).__name__, event=event, correlation=correlation)
    with _lock:
        previous = _workers.get(owner+"/"+name)
        _workers[owner+"/"+name] = Worker(owner=owner, name=name,
            failures=(previous["failures"] if previous else 0)+1,
            deadLetterTransfers=(previous["deadLetterTransfers"] if previous else 0)+int(dead), lastFailure=last)
    logging.warning(json.dumps({"message": "workflow failure", "owner": owner, "worker": name,
                               **last, "deadLetterTransfer": dead}))


def process_workers() -> dict[str, Worker]:
    with _lock:
        return dict(_workers)


def diagnostics(databases: Mapping[str, PostgresContextDatabase]) -> dict[str, object]:
    states: dict[str, object] = {}
    for owner, database in databases.items():
        state: dict[str, object] = {}
        try:
            with database.pool.connection(timeout=3) as connection, connection.transaction():
                connection.execute("SET LOCAL statement_timeout='3s'")
                for name, dispatch, source in (("outbox", "dispatches", "outbox_events"),
                                                ("realtime", "realtime_dispatches", "realtime_publications"),
                                                ("commands", "internal_command_dispatches", "internal_commands")):
                    row = connection.execute(f"""SELECT count(*) AS pending,
                        COALESCE(EXTRACT(EPOCH FROM clock_timestamp()-min(o.created_at)),0)::float8 AS "oldestAgeSeconds",
                        count(*) FILTER (WHERE d.last_error IS NOT NULL) AS failed
                        FROM cafe.{dispatch} d JOIN cafe.{source} o ON o.id=d.event_id
                        WHERE d.completed_at IS NULL""").fetchone()
                    state[name] = row
        except Exception as error:
            failure(owner, "diagnostics", error)
            raise RuntimeError("Diagnostics unavailable") from None
        states[owner] = state
    return {"databases": states, "processWorkers": process_workers(),
            "counterScope": "process lifetime; deadLetterTransfers are confirmed transfers, not queue depths"}
