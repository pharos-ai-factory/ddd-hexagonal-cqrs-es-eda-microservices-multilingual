"""Publish persisted command replies through confirmed RabbitMQ delivery."""

from threading import Event
from typing import TYPE_CHECKING

import pika

from operations.adaptors.diagnostics import failure
from operations.adaptors.messaging.rabbitmq_connection import broker_connection
from operations.adaptors.persistence.dispatch_leases import binary
from operations.adaptors.persistence.command_reply_outbox import claim_reply, finish_reply
from operations.foundation.identity import new_id

if TYPE_CHECKING:
    from operations.adaptors.persistence.postgres_database import PostgresContextDatabase


def relay(database: "PostgresContextDatabase", url: str, stop: Event) -> None:

    while not stop.is_set():
        try:
            with broker_connection(url) as broker:
                channel = broker.channel()
                channel.confirm_delivery()
                while not stop.is_set():
                    token = new_id()
                    row = claim_reply(database, token)
                    if row is None:
                        broker.process_data_events(time_limit=0.1)  # type: ignore[arg-type]
                        continue
                    reason = None
                    try:
                        if row["expired"]:
                            reason = "expired"
                        else:
                            identity = str(row["id"])
                            channel.basic_publish(
                                "cafe.replies",
                                "reply." + database.owner,
                                binary(row["body"]),
                                properties=pika.BasicProperties(
                                    content_type="application/x-protobuf",
                                    delivery_mode=2,
                                    type="reply",
                                    app_id=database.owner,
                                    message_id=identity,
                                    correlation_id=identity,
                                ),
                                mandatory=True,
                            )
                    except Exception:
                        reason = "publication_failed"
                        raise
                    finally:
                        finish_reply(database, row, token, reason)
        except Exception as error:
            failure(database.owner, "replies.reconnect", error)
            stop.wait(1)
