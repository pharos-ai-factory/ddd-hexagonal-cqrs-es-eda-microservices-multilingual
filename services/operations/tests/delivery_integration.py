"""Actual RabbitMQ quarantine evidence for malformed retry metadata."""
import os
from decimal import Decimal
from functools import partial
from concurrent.futures import Future
from hashlib import sha256
from threading import Event, Thread
import time
import pika
import pytest
from operations.adaptors.codec import drinks_ready, order_placed
from operations.adaptors.delivery import EventSubscription, binary, broker_connection, consume
from operations.adaptors.generated.cafe.v1.events_pb2 import Event as WireEvent
from operations.foundation.identity import derived_id, new_id
from operations.contracts.events import DrinksReady, OrderPlaced
from operations.foundation.application import Metadata, Outcome
from operations.adaptors.postgres import fingerprint


@pytest.mark.parametrize("counter", ["invalid", -2147483648, 2147483647, Decimal("1.5"), True, None])
def test_malformed_retry_header_is_quarantined_without_reconnect_livelock(counter: object) -> None:
    consumer = "collection.test-"+new_id()
    queue = "ref."+consumer
    stop, handled = Event(), Event()
    connection = broker_connection(os.environ["BROKER_ADMIN_URL"])
    channel = connection.channel()
    worker = None
    poison_id = new_id()
    try:
        channel.confirm_delivery()
        for suffix in ("", ".retry", ".dead"):
            channel.queue_declare(queue+suffix, durable=True,
                                  arguments={"x-queue-type": "quorum", "x-delivery-limit": -1})
            channel.queue_bind(queue+suffix, "ref.collection.delivery", queue+suffix)
        channel.basic_publish("ref.collection.delivery", queue, b"\x00", mandatory=True,
                              properties=pika.BasicProperties(delivery_mode=2, message_id=poison_id,
                                                              headers={"ref-attempt": counter}))
        valid = WireEvent(id=new_id(), name="preparation.drinks-ready", context="preparation",
                          visibility="integration", contract_version=1, aggregate_kind="ticket",
                          aggregate_id=new_id(), aggregate_version=3, correlation_id=new_id(),
                          causation_id=new_id(), occurred_at="2026-09-27T12:00:00Z")
        valid.drinks_ready.order_id, valid.drinks_ready.customer_id = new_id(), new_id()
        channel.basic_publish("ref.collection.delivery", queue, valid.SerializeToString(), mandatory=True,
                              properties=pika.BasicProperties(content_type="application/x-protobuf",
                                  delivery_mode=2, message_id=valid.id, type=valid.name, app_id=valid.context,
                                  correlation_id=valid.correlation_id, headers={"contract-version": 1}))

        def handle(metadata: Metadata, payload: DrinksReady) -> Outcome:
            handled.set()
            return {"aggregateId": metadata.target, "version": 1, "status": "ready"}

        subscription = EventSubscription[DrinksReady]("collection", consumer, valid.name, drinks_ready,
                                                lambda event: event["orderId"], handle)
        worker = Thread(target=partial(consume, os.environ["COLLECTION_BROKER_URL"], subscription, stop), daemon=True)
        worker.start()
        deadline = time.monotonic()+5
        dead: tuple[pika.BasicProperties, bytes] | None = None
        while time.monotonic() < deadline:
            method, properties, body = channel.basic_get(queue+".dead", auto_ack=False)
            if method:
                assert properties is not None and method.delivery_tag is not None
                # Pika returns bytes; its third-party stub incorrectly declares str.
                dead = (properties, binary(body))
                channel.basic_ack(method.delivery_tag)
                break
            time.sleep(0.05)
        assert dead is not None, "Malformed retry metadata must reach quarantine, not loop on reconnect"
        assert dead[0].message_id == poison_id and dead[1] == b"\x00"
        assert dead[0].headers is not None and type(dead[0].headers["ref-attempt"]) is int
        assert handled.wait(5), "A poison delivery must not block the next valid message"
    finally:
        stop.set()
        if worker:
            worker.join(timeout=7)
        for suffix in ("", ".retry", ".dead"):
            channel.queue_delete(queue+suffix)
        connection.close()


def test_typed_order_delivery_preserves_original_receipt_material() -> None:
    consumer = "preparation.test-"+new_id()
    queue, stop = "ref."+consumer, Event()
    received: Future[tuple[Metadata, OrderPlaced]] = Future()
    event = WireEvent(id=new_id(), name="ordering.order-placed", context="ordering", visibility="integration",
        contract_version=1, aggregate_kind="order", aggregate_id=new_id(), aggregate_version=3,
        correlation_id=new_id(), causation_id=new_id(), occurred_at="2026-09-27T12:00:00Z")
    order = event.order_placed
    order.order_id, order.customer_id, order.edition_id, order.currency = event.aggregate_id, new_id(), new_id(), "EUR"
    line_id = new_id()
    order.lines.add(id=line_id, offer_code="C1", name="Coffee", quantity=2, minor=350)
    body = event.SerializeToString(deterministic=True)
    wire_payload = {"orderId": order.order_id, "customerId": order.customer_id, "editionId": order.edition_id,
        "currency": "EUR", "lines": [{"id": line_id, "offerCode": "C1", "name": "Coffee", "quantity": 2, "minor": "350"}]}

    def handle(metadata: Metadata, payload: OrderPlaced) -> Outcome:
        received.set_result((metadata, payload))
        return {"aggregateId": metadata.target, "version": 1, "status": "queued"}

    subscription = EventSubscription[OrderPlaced]("preparation", consumer, event.name, order_placed,
        lambda payload: derived_id("ticket", payload["orderId"]), handle)
    worker = Thread(target=partial(consume, os.environ["PREPARATION_BROKER_URL"], subscription, stop), daemon=True)
    with broker_connection(os.environ["BROKER_ADMIN_URL"]) as connection:
        channel = connection.channel()
        try:
            channel.confirm_delivery()
            channel.queue_declare(queue, durable=True, arguments={"x-queue-type": "quorum", "x-delivery-limit": -1})
            channel.queue_bind(queue, "ref.preparation.delivery", queue)
            channel.basic_publish("ref.preparation.delivery", queue, body, mandatory=True,
                properties=pika.BasicProperties(content_type="application/x-protobuf", delivery_mode=2,
                    message_id=event.id, type=event.name, app_id=event.context,
                    correlation_id=event.correlation_id, headers={"contract-version": 1}))
            worker.start()
            metadata, payload = received.result(timeout=5)
            assert payload["lines"][0]["minor"] == 350
            assert metadata.input == wire_payload
            assert metadata.source_hash == sha256(body).hexdigest()
            assert metadata.id == derived_id(consumer, event.id)
            assert fingerprint({"expected": metadata.expected, "input": metadata.input}) == fingerprint(
                {"expected": None, "input": wire_payload})
        finally:
            stop.set()
            if worker.ident is not None:
                worker.join(timeout=7)
            channel.queue_delete(queue)
