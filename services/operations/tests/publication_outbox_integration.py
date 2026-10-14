"""Real atomic publication receipts, independent subscribers and lost-ACK recovery."""

from operations.adaptors.generated.message_names import PreparationDomainEvent, PreparationSubscription

from concurrent.futures import ThreadPoolExecutor
from dataclasses import replace
from hashlib import sha256
import os
from urllib.parse import urlsplit, urlunsplit
from uuid import uuid4
from threading import Event
from unittest.mock import patch
import pika
import psycopg
import pytest
from operations.adaptors.messaging.event_codec import decode
from operations.adaptors.persistence.dispatch_leases import binary
from operations.adaptors.messaging.rabbitmq_connection import broker_connection
from operations.adaptors.persistence.integration_event_outbox import PostgresPublicationOutbox
from operations.adaptors.generated.event_preparation import publish_DrinksReady
from operations.apps.composition.preparation import PreparationContainer, requests, subscriptions
from operations.adaptors.messaging.rabbitmq_consumer import EventSubscription
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_requests_pb2 import (
    Request,
)
from operations.contexts.preparation.application.commands.accept_order import AcceptOrderCommand
from operations.contexts.preparation.adaptors.messaging.private_events import DRINKS_READY_DOMAIN
from operations.contracts.events.generated.preparation import DrinksReadyIntegrationEvent
from operations.foundation.application import Metadata
from operations.foundation.identity import derived_id


def test_publication_receipt_outbox_and_subscription_recovery() -> None:
    assert os.environ["CAFE_DISPOSABLE_PROJECT"].startswith("cafe-reference-test-")
    container = PreparationContainer()
    container.config.database_url.from_value(os.environ["PREPARATION_DATABASE_URL"])
    container.init_resources()
    db = container.database()
    target, order, customer, correlation = (str(uuid4()) for _ in range(4))
    consumer = PreparationSubscription.PUBLISH_DRINKS_READY
    try:
        accept = next(worker for worker in subscriptions(container) if worker.command)
        source_id = str(uuid4())
        command = Metadata(
            id=derived_id(accept.consumer, source_id),
            target=target,
            name=accept.consumer,
            correlation=correlation,
            consumer=accept.consumer,
            source_id=source_id,
            source_hash="a" * 64,
            input={"orderId": order},
        )
        accept_command = AcceptOrderCommand(orderId=order, customerId=customer, instructions="1 × Tea")
        queued = container.codec().encode(command, accept_command)

        def deliver(_url: str, subscription: EventSubscription[AcceptOrderCommand], _stop: Event) -> None:
            assert subscription.command_decoder is not None
            metadata, value = subscription.command_decoder(queued)
            assert subscription.handle(metadata, value)["status"] == "queued"

        with patch("operations.apps.runtime.consume", deliver):
            accept.run("unused", Event())
        registry = requests(container)
        for version, action, status in [
            (1, "start_preparation", "preparing"),
            (2, "complete_preparation", "ready"),
        ]:
            request = Request(contract_version=1, request_id=str(uuid4()), context="preparation")
            request.command.metadata.command_id = str(uuid4())
            request.command.metadata.aggregate_id = target
            request.command.metadata.correlation_id = correlation
            request.command.metadata.expected_version = version
            getattr(request.command, action).SetInParent()
            assert registry.handle(request).outcome.status == status
        with db.pool.connection() as connection:
            source = connection.execute(
                "SELECT body FROM cafe.outbox_events WHERE aggregate_id=%s AND event_name=%s",
                (target, PreparationDomainEvent.DRINKS_READY),
            ).fetchone()
            assert source
            body = binary(source["body"])
            assert connection.execute(
                "SELECT count(*) AS n FROM cafe.outbox_events WHERE aggregate_id=%s AND visibility='integration'",
                (target,),
            ).fetchone() == {"n": 0}
        event, payload = decode(body)
        fact = DRINKS_READY_DOMAIN.parse(event)
        incoming = Metadata(
            id=derived_id(consumer, event.id),
            target=target,
            name=consumer,
            correlation=correlation,
            consumer=consumer,
            source_id=event.id,
            source_hash=sha256(body).hexdigest(),
            input=payload,
            causation=event.id,
            source_kind=event.aggregate_kind,
            source_version=event.aggregate_version,
            source_time=event.occurred_at,
        )
        assert incoming.source_version == 3
        address = urlsplit(os.environ["DATABASE_ADMIN_URL"])
        admin_url = urlunsplit(address._replace(path="/cafe_preparation"))
        with psycopg.connect(admin_url, autocommit=True) as admin:
            admin.execute(
                f"ALTER TABLE cafe.consumer_receipts ADD CONSTRAINT reject_publication_probe CHECK(event_id <> '{event.id}') NOT VALID"
            )
            try:
                with pytest.raises(psycopg.errors.CheckViolation):
                    container.publish_drinks_ready().handle(incoming, fact)
                assert admin.execute(
                    "SELECT count(*) FROM cafe.outbox_events WHERE aggregate_id=%s AND visibility='integration'",
                    (target,),
                ).fetchone() == (0,)
                assert admin.execute(
                    "SELECT count(*) FROM cafe.consumer_receipts WHERE event_id=%s", (event.id,)
                ).fetchone() == (0,)
            finally:
                admin.execute("ALTER TABLE cafe.consumer_receipts DROP CONSTRAINT reject_publication_probe")
            with ThreadPoolExecutor(max_workers=2) as pool:
                results = list(
                    pool.map(lambda _: container.publish_drinks_ready().handle(incoming, fact), range(2))
                )
            assert results[0] == results[1]
            first = admin.execute(
                "SELECT body FROM cafe.outbox_events WHERE aggregate_id=%s AND visibility='integration'",
                (target,),
            ).fetchall()
            assert len(first) == 1
            public, _ = decode(bytes(first[0][0]))
            assert (
                public.aggregate_version,
                public.causation_id,
                public.correlation_id,
                public.occurred_at,
            ) == (3, event.id, correlation, event.occurred_at)
            assert admin.execute(
                "SELECT count(*) FROM cafe.dispatches d JOIN cafe.outbox_events e ON e.id=d.event_id WHERE e.aggregate_id=%s AND e.visibility='integration'",
                (target,),
            ).fetchone() == (1,)
            with pytest.raises(ValueError, match="conflicting"):
                container.publish_drinks_ready().handle(replace(incoming, source_hash="b" * 64), fact)
            assert admin.execute("SELECT version FROM cafe.aggregates WHERE id=%s", (target,)).fetchone() == (
                3,
            )
            # One receipt never suppresses another independent subscription to the same fact.
            audit = "preparation.audit-ready"
            other = PostgresPublicationOutbox[DrinksReadyIntegrationEvent](
                db, audit, "ticket", publish_DrinksReady
            )
            other.publish(
                replace(incoming, consumer=audit, name=audit, id=derived_id(audit, event.id)),
                DrinksReadyIntegrationEvent(orderId=order, customerId=customer),
            )
            assert admin.execute(
                "SELECT count(*) FROM cafe.consumer_receipts WHERE event_id=%s", (event.id,)
            ).fetchone() == (2,)
            with broker_connection(os.environ["BROKER_ADMIN_URL"]) as broker:
                channel = broker.channel()
                channel.confirm_delivery()
                queue = "ref." + consumer
                observer = channel.queue_declare(queue="", exclusive=True).method.queue
                assert observer
                channel.queue_bind(observer, "cafe.events", "domain." + event.name)
                channel.basic_publish(
                    "cafe.events",
                    "domain." + event.name,
                    body,
                    mandatory=True,
                    properties=pika.BasicProperties(
                        content_type="application/x-protobuf",
                        delivery_mode=2,
                        message_id=event.id,
                        type=event.name,
                        app_id="preparation",
                        correlation_id=correlation,
                        headers={"contract-version": 1},
                    ),
                )
                delivery, _, received = channel.basic_get(queue, auto_ack=False)
                assert delivery and binary(received) == body
                container.publish_drinks_ready().handle(incoming, fact)
                channel.close()  # The database committed, but this subscriber's ACK never reached RabbitMQ.
                channel = broker.channel()
                delivery, _, received = channel.basic_get(queue, auto_ack=False)
                assert delivery and delivery.redelivered and binary(received) == body
                container.publish_drinks_ready().handle(incoming, fact)
                assert delivery.delivery_tag is not None
                channel.basic_ack(delivery.delivery_tag)
                independent, _, observed = channel.basic_get(observer, auto_ack=False)
                assert independent and binary(observed) == body
                assert independent.delivery_tag is not None
                channel.basic_ack(independent.delivery_tag)
                assert channel.basic_get(queue, auto_ack=False)[0] is None
            final = admin.execute(
                "SELECT body FROM cafe.outbox_events WHERE aggregate_id=%s AND visibility='integration' AND causation_id=%s ORDER BY id",
                (target, event.id),
            ).fetchall()
            assert len(final) == 2
            assert bytes(first[0][0]) in [bytes(row[0]) for row in final]
    finally:
        container.shutdown_resources()
