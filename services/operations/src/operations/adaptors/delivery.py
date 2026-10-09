"""Fenced outbox dispatch and confirmed, manual-acknowledgement consumption."""
from operations.foundation.secrets import secret
import base64
from collections.abc import Callable
from dataclasses import dataclass
from hashlib import sha256
import json
import os
from threading import Event
from typing import Literal, NotRequired, TypedDict
import urllib.request
import pika
from pika.adapters.blocking_connection import BlockingChannel, BlockingConnection
from pika.spec import Basic
from operations.adaptors.codec import decode
from operations.adaptors.diagnostics import failure
from operations.adaptors.generated.cafe.v1.events_pb2 import Event as WireEvent
from operations.adaptors.postgres import PostgresContextDatabase
from operations.foundation.application import Metadata, Outcome
from operations.foundation.domain import Rejection, identifier, integer, text
from operations.foundation.identity import derived_id, new_id


class OutboxDispatch(TypedDict):
    """Carries immutable publication bytes and the fenced lease used to complete one dispatch."""
    id: str
    token: str
    generation: int
    body: bytes
    channel: NotRequired[str]


def binary(value: object) -> bytes:
    if not isinstance(value, (bytes, bytearray, memoryview)):
        raise ValueError("Publication bytes are required")
    return bytes(value)


def claim(database: PostgresContextDatabase, realtime: bool | Literal["commands"] = False) -> OutboxDispatch | None:
    dispatch = "internal_command_dispatches" if realtime == "commands" else "realtime_dispatches" if realtime else "dispatches"
    source = "internal_commands" if realtime == "commands" else "realtime_publications" if realtime else "outbox_events"
    token = new_id()
    with database.pool.connection() as connection:
        row = connection.execute(f"""WITH candidate AS (
            SELECT event_id FROM cafe.{dispatch} WHERE completed_at IS NULL
            AND available_at<=clock_timestamp() AND (lease_until IS NULL OR lease_until<clock_timestamp())
            ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
        ), claimed AS (
            UPDATE cafe.{dispatch} d SET lease_token=%s,lease_until=clock_timestamp()+interval '30 seconds',
            generation=generation+1 FROM candidate c WHERE d.event_id=c.event_id
            RETURNING d.event_id,d.generation
        ) SELECT o.*,c.generation FROM claimed c JOIN cafe.{source} o ON o.id=c.event_id""", (token,)).fetchone()
    if not row:
        return None
    result = OutboxDispatch(id=identifier(str(row["id"])), token=token,
                      generation=integer(row["generation"]), body=binary(row["body"]))
    if realtime is True:
        result["channel"] = text(row["channel"])
    return result


def finish(database: PostgresContextDatabase, row: OutboxDispatch, realtime: bool | Literal["commands"] = False, error: str | None = None) -> None:
    table = "internal_command_dispatches" if realtime == "commands" else "realtime_dispatches" if realtime else "dispatches"
    with database.pool.connection() as connection:
        connection.execute(f"""UPDATE cafe.{table} SET lease_token=NULL,lease_until=NULL,
            available_at=clock_timestamp()+interval '1 second',
            completed_at=CASE WHEN %s::text IS NULL THEN clock_timestamp() ELSE NULL END,last_error=%s
            WHERE event_id=%s AND lease_token=%s AND generation=%s AND lease_until>clock_timestamp()""",
            (error, error, row["id"], row["token"], row["generation"]))


def broker_connection(url: str) -> BlockingConnection:
    parameters = pika.URLParameters(url)
    parameters.heartbeat = 30
    parameters.blocked_connection_timeout = 5
    parameters.socket_timeout = 5
    return pika.BlockingConnection(parameters)


def relay(database: PostgresContextDatabase, broker_url: str, stop: Event,
          command_header: Callable[[bytes], tuple[str, str, str]] | None = None) -> None:
    while not stop.is_set():
        try:
            with broker_connection(broker_url) as connection:
                channel = connection.channel()
                channel.confirm_delivery()
                while not stop.is_set():
                    row = claim(database, "commands" if command_header else False)
                    if not row:
                        # Pika accepts fractional seconds; types-pika incorrectly declares int.
                        connection.process_data_events(time_limit=0.1)  # type: ignore[arg-type]
                        continue
                    try:
                        if command_header:
                            identity, name, correlation = command_header(row["body"])
                            exchange, route, owner = "ref."+database.owner+".delivery", "ref."+name, database.owner
                        else:
                            event, _ = decode(row["body"])
                            identity, name, correlation = event.id, event.name, event.correlation_id
                            exchange, route, owner = "cafe.events", event.visibility+"."+event.name, event.context
                        channel.basic_publish(exchange, route, row["body"],
                            properties=pika.BasicProperties(content_type="application/x-protobuf", delivery_mode=2,
                                message_id=identity, type=name, app_id=owner,
                                correlation_id=correlation, headers={"contract-version": 1}), mandatory=True)
                        finish(database, row, "commands" if command_header else False)
                    except Exception as error:
                        failure(database.owner, "outbox.publish", error, row["id"])
                        finish(database, row, "commands" if command_header else False, error=type(error).__name__)
                        raise
        except Exception as error:
            failure(database.owner, "outbox.reconnect", error)
            stop.wait(1)


def realtime_relay(database: PostgresContextDatabase, stop: Event) -> None:
    while not stop.is_set():
        row = None
        try:
            row = claim(database, True)
            if not row:
                stop.wait(0.1)
                continue
            request = urllib.request.Request(os.environ["REALTIME_GATEWAY_URL"]+"/api/publish", method="POST",
                data=json.dumps({"channel": row["channel"], "b64data": base64.b64encode(row["body"]).decode(),
                                 "idempotency_key": str(row["id"])}).encode(),
                headers={"Content-Type": "application/json", "X-API-Key": secret(database.owner.upper()+"_REALTIME_KEY"),
                         "X-Centrifugo-Error-Mode": "transport"})
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


@dataclass(frozen=True)
class EventSubscription[P]:
    """Binds a stable owner consumer identity to decoding, target selection and application handling."""
    owner: str
    consumer: str
    event: str
    parse: Callable[[WireEvent], P]
    target: Callable[[P], str]
    handle: Callable[[Metadata, P], Outcome]
    command_decoder: Callable[[bytes], tuple[Metadata, P]] | None = None


def consume[P](url: str, subscription: EventSubscription[P], stop: Event) -> None:
    sub = subscription
    queue = "ref."+sub.consumer+(".command" if sub.command_decoder else "")
    while not stop.is_set():
        try:
            with broker_connection(url) as connection:
                channel = connection.channel()
                channel.confirm_delivery()
                channel.basic_qos(prefetch_count=1)

                def receive(channel: BlockingChannel, method: Basic.Deliver,
                            properties: pika.BasicProperties, body: bytes) -> None:
                    if method.delivery_tag is None:
                        raise ValueError("Missing broker delivery tag")
                    validation = True
                    event_id, correlation = "", ""
                    counter = (properties.headers or {}).get("ref-attempt", 0)
                    valid_attempt = type(counter) is int and 0 <= counter <= 3
                    attempt = counter if valid_attempt else 0
                    try:
                        if not valid_attempt:
                            raise ValueError("Invalid retry counter")
                        if sub.command_decoder:
                            metadata, payload = sub.command_decoder(body)
                            if (properties.content_type != "application/x-protobuf" or properties.delivery_mode != 2
                                or properties.message_id != metadata.id or properties.type != sub.consumer+".command"
                                or properties.app_id != sub.owner or properties.correlation_id != metadata.correlation
                                or (properties.headers or {}).get("contract-version") != 1):
                                raise ValueError("Invalid command properties")
                            event_id, correlation = metadata.id, metadata.correlation
                        else:
                            event, wire_payload = decode(body)
                            event_id, correlation = event.id, event.correlation_id
                            if (event.name != sub.event or (event.visibility == "domain" and event.context != sub.owner)
                                or properties.content_type != "application/x-protobuf" or properties.delivery_mode != 2
                                or properties.message_id != event.id or properties.type != event.name
                                or properties.app_id != event.context or properties.correlation_id != event.correlation_id
                                or (properties.headers or {}).get("contract-version") != 1):
                                raise ValueError("AMQP metadata disagrees with the event")
                            payload = sub.parse(event)
                            metadata = Metadata(id=derived_id(sub.consumer, event.id), target=sub.target(payload), name=sub.consumer,
                                correlation=event.correlation_id, input=wire_payload, causation=event.id, consumer=sub.consumer,
                                source_id=event.id, source_hash=sha256(body).hexdigest())
                        validation = False
                        outcome = sub.handle(metadata, payload)
                        if "rejection" in outcome:
                            raise Rejection(**outcome["rejection"])
                    except Exception as error:
                        headers = {k: v for k, v in (properties.headers or {}).items() if k != "x-death"}
                        suffix = ".dead" if validation or isinstance(error, Rejection) or attempt >= 3 else ".retry"
                        headers.update({"ref-attempt": attempt+1, "ref-failure": type(error).__name__})
                        properties.headers = headers
                        # Confirmation of the transfer precedes source acknowledgement.
                        channel.basic_publish("ref."+sub.owner+".delivery", queue+suffix, body,
                                              properties=properties, mandatory=True)
                        failure(sub.owner, sub.consumer, error, event_id, correlation, suffix == ".dead")
                    channel.basic_ack(method.delivery_tag)

                channel.basic_consume(queue, receive, auto_ack=False)
                while not stop.is_set():
                    # Pika accepts fractional seconds; types-pika incorrectly declares int.
                    connection.process_data_events(time_limit=0.5)  # type: ignore[arg-type]
        except Exception as error:
            failure(sub.owner, sub.consumer+".reconnect", error)
            stop.wait(1)
