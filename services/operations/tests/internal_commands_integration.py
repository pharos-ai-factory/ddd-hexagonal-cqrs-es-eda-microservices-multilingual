"""Real receiving commits, immutable command bytes and independent queue recovery."""
from dataclasses import replace
from threading import Event, Thread
from time import monotonic, sleep
from uuid import uuid4
import os
from urllib.parse import urlsplit, urlunsplit
import psycopg
import pytest
from operations.apps.composition.preparation import command_codec, command_header, restore
from operations.adaptors.internal_commands import PostgresDurableCommandOutbox
from operations.adaptors.delivery import EventSubscription, binary, broker_connection, claim, consume, finish, relay
from operations.adaptors.postgres import PostgresContextDatabase, PostgresAggregateCommandStore, PostgresAggregateQueries
from operations.adaptors.generated.cafe.v1.events_pb2 import Event as WireEvent
from operations.contexts.preparation.application import AcceptOrderCommand, AcceptOrderCommandHandler
from operations.foundation.application import Metadata, Outcome
from operations.foundation.identity import derived_id


def test_internal_command_acceptance_and_recovery() -> None:
    assert os.environ["CAFE_DISPOSABLE_PROJECT"].startswith("cafe-reference-test-")
    database = PostgresContextDatabase("preparation", os.environ["PREPARATION_DATABASE_URL"])
    address = urlsplit(os.environ["DATABASE_ADMIN_URL"])
    admin_url = urlunsplit(address._replace(path="/cafe_preparation"))
    codec = command_codec()
    outbox = PostgresDurableCommandOutbox(database, codec)
    source, target, correlation = (str(uuid4()) for _ in range(3))
    command = AcceptOrderCommand(orderId=source, customerId=correlation, instructions="1 × Coffee")
    metadata = Metadata(id=derived_id(codec.consumer, source), target=target, name=codec.consumer,
        correlation=correlation, causation=source, consumer=codec.consumer, source_id=source, source_hash="a"*64,
        input={"source": "original event receipt"})
    handler = AcceptOrderCommandHandler(PostgresAggregateCommandStore(database, "ticket", restore))
    queries = PostgresAggregateQueries(database, "ticket", restore)
    stop = Event()
    workers: list[Thread] = []
    attempts, completed = 0, 0
    unavailable = True

    def execute(incoming: Metadata, value: AcceptOrderCommand) -> Outcome:
        nonlocal attempts, completed
        attempts += 1
        if unavailable:
            raise OSError("injected infrastructure outage")
        outcome = handler.execute(incoming, value)
        completed += 1
        return outcome

    def unused(_: WireEvent) -> AcceptOrderCommand:
        raise AssertionError("Private command decoder required")

    with psycopg.connect(admin_url, autocommit=True) as admin, broker_connection(os.environ["BROKER_ADMIN_URL"]) as connection:
        channel = connection.channel()
        channel.confirm_delivery()
        queue = "ref."+codec.consumer+".command"
        try:
            admin.execute(f"ALTER TABLE cafe.internal_command_dispatches ADD CONSTRAINT reject_handoff_probe CHECK(event_id <> '{metadata.id}') NOT VALID")
            with pytest.raises(psycopg.errors.CheckViolation):
                outbox.enqueue(metadata, command)
            assert admin.execute("SELECT count(*) FROM cafe.internal_commands WHERE id=%s", (metadata.id,)).fetchone() == (0,)
            admin.execute("ALTER TABLE cafe.internal_command_dispatches DROP CONSTRAINT reject_handoff_probe")
            outbox.enqueue(metadata, command)
            outbox.enqueue(metadata, command)
            with pytest.raises(ValueError):
                outbox.enqueue(replace(metadata, source_hash="b"*64), command)
            assert queries.get(target) is None
            abandoned = claim(database, "commands")
            assert abandoned and abandoned["id"] == metadata.id
            admin.execute("UPDATE cafe.internal_command_dispatches SET lease_until=clock_timestamp()-interval '1 second' WHERE event_id=%s", (metadata.id,))
            recovered = claim(database, "commands")
            assert recovered and recovered["body"] == abandoned["body"]
            finish(database, abandoned, "commands")
            assert admin.execute("SELECT completed_at FROM cafe.internal_command_dispatches WHERE event_id=%s", (metadata.id,)).fetchone() == (None,)
            finish(database, recovered, "commands", "recover")
            subscription = EventSubscription("preparation", codec.consumer, "ordering.order-placed", unused,
                lambda _: target, execute, codec.decode)
            workers = [Thread(target=consume, args=(os.environ["PREPARATION_BROKER_URL"], subscription, stop)),
                       Thread(target=relay, args=(database, os.environ["PREPARATION_BROKER_URL"], stop, command_header))]
            for worker in workers:
                worker.start()
            deadline = monotonic()+12
            while True:
                method, properties, body = channel.basic_get(queue+".dead", auto_ack=False)
                if method:
                    break
                assert monotonic() < deadline, "Command did not reach bounded quarantine"
                sleep(0.03)
            wire_body = binary(body)
            assert attempts == 4 and properties and wire_body == recovered["body"]
            assert queries.get(target) is None
            unavailable = False
            properties.headers = {"contract-version": 1}
            channel.basic_publish("ref.preparation.delivery", queue, wire_body, properties=properties, mandatory=True)
            assert method.delivery_tag is not None
            channel.basic_ack(method.delivery_tag)
            while completed < 1:
                assert monotonic() < deadline
                sleep(0.03)
            channel.basic_publish("ref.preparation.delivery", queue, wire_body, properties=properties, mandatory=True)
            while completed < 2:
                assert monotonic() < deadline
                sleep(0.03)
            current = queries.get(target)
            assert current and current["version"] == 1
            assert current["state"]["instructions"] == command["instructions"]
            with database.pool.connection() as runtime:
                with pytest.raises(psycopg.errors.InsufficientPrivilege):
                    runtime.execute("UPDATE cafe.internal_commands SET body=%s WHERE id=%s", (wire_body, metadata.id))
        finally:
            stop.set()
            for worker in workers:
                worker.join(timeout=6)
                assert not worker.is_alive()
            admin.execute("ALTER TABLE cafe.internal_command_dispatches DROP CONSTRAINT IF EXISTS reject_handoff_probe")
            database.pool.close()
