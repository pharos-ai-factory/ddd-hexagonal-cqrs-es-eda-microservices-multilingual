"""Publish persisted event or command bytes through confirmed RabbitMQ delivery."""

from collections.abc import Callable
from threading import Event

import pika

from operations.adaptors.diagnostics import failure
from operations.adaptors.messaging.event_codec import decode
from operations.adaptors.messaging.rabbitmq_connection import broker_connection
from operations.adaptors.persistence.dispatch_leases import claim, finish
from operations.adaptors.persistence.postgres_database import PostgresContextDatabase


def relay(
    database: PostgresContextDatabase,
    broker_url: str,
    stop: Event,
    command_header: Callable[[bytes], tuple[str, str, str]] | None = None,
) -> None:
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
                            exchange, route, owner = (
                                "ref." + database.owner + ".delivery",
                                "ref." + name,
                                database.owner,
                            )
                        else:
                            event, _ = decode(row["body"])
                            identity, name, correlation = event.id, event.name, event.correlation_id
                            exchange, route, owner = (
                                "cafe.events",
                                event.visibility + "." + event.name,
                                event.context,
                            )
                        channel.basic_publish(
                            exchange,
                            route,
                            row["body"],
                            properties=pika.BasicProperties(
                                content_type="application/x-protobuf",
                                delivery_mode=2,
                                message_id=identity,
                                type=name,
                                app_id=owner,
                                correlation_id=correlation,
                                headers={"contract-version": 1},
                            ),
                            mandatory=True,
                        )
                        finish(database, row, "commands" if command_header else False)
                    except Exception as error:
                        failure(database.owner, "outbox.publish", error, row["id"])
                        finish(
                            database, row, "commands" if command_header else False, error=type(error).__name__
                        )
                        raise
        except Exception as error:
            failure(database.owner, "outbox.reconnect", error)
            stop.wait(1)
