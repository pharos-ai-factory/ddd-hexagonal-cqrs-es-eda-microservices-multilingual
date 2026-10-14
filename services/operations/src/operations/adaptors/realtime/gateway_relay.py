"""Publish persisted realtime bytes through the authorised gateway and complete their leases."""

import base64
import json
import os
import urllib.request
from threading import Event

from operations.adaptors.diagnostics import failure
from operations.adaptors.persistence.dispatch_leases import claim, finish
from operations.adaptors.persistence.postgres_database import PostgresContextDatabase
from operations.foundation.secrets import secret


def relay(database: PostgresContextDatabase, stop: Event) -> None:
    while not stop.is_set():
        row = None
        try:
            row = claim(database, True)
            if not row:
                stop.wait(0.1)
                continue
            request = urllib.request.Request(
                os.environ["REALTIME_GATEWAY_URL"] + "/api/publish",
                method="POST",
                data=json.dumps(
                    {
                        "channel": row["channel"],
                        "b64data": base64.b64encode(row["body"]).decode(),
                        "idempotency_key": str(row["id"]),
                    }
                ).encode(),
                headers={
                    "Content-Type": "application/json",
                    "X-API-Key": secret(database.owner.upper() + "_REALTIME_KEY"),
                    "X-Centrifugo-Error-Mode": "transport",
                },
            )
            with urllib.request.urlopen(request, timeout=5) as response:
                if "error" in json.load(response):
                    raise RuntimeError("Centrifugo rejected publication")
            finish(database, row, True)
        except Exception as error:
            if row:
                try:
                    finish(database, row, True, type(error).__name__)
                except Exception as finish_error:
                    # A lost database connection also leaves a recoverable lease.
                    failure(database.owner, "realtime.complete", finish_error, row["id"])
            failure(database.owner, "realtime.publish", error, row["id"] if row else "")
            stop.wait(1)
