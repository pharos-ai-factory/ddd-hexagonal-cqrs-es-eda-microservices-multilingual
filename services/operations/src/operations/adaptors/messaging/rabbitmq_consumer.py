"""Consume durable events and commands with bounded retries and manual acknowledgements."""

from collections.abc import Callable
from dataclasses import dataclass
from hashlib import sha256
from threading import Event

import pika
from pika.adapters.blocking_connection import BlockingChannel
from pika.spec import Basic

from operations.adaptors.diagnostics import failure
from operations.adaptors.generated.cafe.v1.events_pb2 import Event as WireEvent
from operations.adaptors.messaging.event_codec import decode
from operations.adaptors.messaging.rabbitmq_connection import broker_connection
from operations.foundation.application import Metadata, Outcome
from operations.foundation.domain import Rejection
from operations.foundation.identity import derived_id


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
    source_target: bool = False


def consume[P](url: str, subscription: EventSubscription[P], stop: Event) -> None:
    sub = subscription
    queue = "ref." + sub.consumer + (".command" if sub.command_decoder else "")
    while not stop.is_set():
        try:
            with broker_connection(url) as connection:
                channel = connection.channel()
                channel.confirm_delivery()
                channel.basic_qos(prefetch_count=1)

                def receive(
                    channel: BlockingChannel,
                    method: Basic.Deliver,
                    properties: pika.BasicProperties,
                    body: bytes,
                ) -> None:
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
                            if (
                                properties.content_type != "application/x-protobuf"
                                or properties.delivery_mode != 2
                                or properties.message_id != metadata.id
                                or properties.type != sub.consumer + ".command"
                                or properties.app_id != sub.owner
                                or properties.correlation_id != metadata.correlation
                                or (properties.headers or {}).get("contract-version") != 1
                            ):
                                raise ValueError("Invalid command properties")
                            event_id, correlation = metadata.id, metadata.correlation
                        else:
                            event, wire_payload = decode(body)
                            event_id, correlation = event.id, event.correlation_id
                            if (
                                event.name != sub.event
                                or (event.visibility == "domain" and event.context != sub.owner)
                                or properties.content_type != "application/x-protobuf"
                                or properties.delivery_mode != 2
                                or properties.message_id != event.id
                                or properties.type != event.name
                                or properties.app_id != event.context
                                or properties.correlation_id != event.correlation_id
                                or (properties.headers or {}).get("contract-version") != 1
                            ):
                                raise ValueError("AMQP metadata disagrees with the event")
                            payload = sub.parse(event)
                            metadata = Metadata(
                                id=derived_id(sub.consumer, event.id),
                                target=event.aggregate_id if sub.source_target else sub.target(payload),
                                name=sub.consumer,
                                correlation=event.correlation_id,
                                input=wire_payload,
                                causation=event.id,
                                consumer=sub.consumer,
                                source_id=event.id,
                                source_hash=sha256(body).hexdigest(),
                                source_kind=event.aggregate_kind,
                                source_version=event.aggregate_version,
                                source_time=event.occurred_at,
                            )
                        validation = False
                        outcome = sub.handle(metadata, payload)
                        if "rejection" in outcome:
                            raise Rejection(**outcome["rejection"])
                    except Exception as error:
                        headers = {k: v for k, v in (properties.headers or {}).items() if k != "x-death"}
                        suffix = (
                            ".dead"
                            if validation or isinstance(error, Rejection) or attempt >= 3
                            else ".retry"
                        )
                        headers.update({"ref-attempt": attempt + 1, "ref-failure": type(error).__name__})
                        properties.headers = headers
                        # Confirmation of the transfer precedes source acknowledgement.
                        channel.basic_publish(
                            "ref." + sub.owner + ".delivery",
                            queue + suffix,
                            body,
                            properties=properties,
                            mandatory=True,
                        )
                        failure(sub.owner, sub.consumer, error, event_id, correlation, suffix == ".dead")
                    channel.basic_ack(method.delivery_tag)

                channel.basic_consume(queue, receive, auto_ack=False)
                while not stop.is_set():
                    # Pika accepts fractional seconds; types-pika incorrectly declares int.
                    connection.process_data_events(time_limit=0.5)  # type: ignore[arg-type]
        except Exception as error:
            failure(sub.owner, sub.consumer + ".reconnect", error)
            stop.wait(1)
