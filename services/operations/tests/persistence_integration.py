"""Real PostgreSQL evidence; invoked explicitly by the isolated integration lane."""
from dataclasses import replace
from concurrent.futures import ThreadPoolExecutor
import os
import pytest
from operations.adaptors.postgres import Database, Commands, Queries, Row
from operations.adaptors.delivery import claim, finish
from operations.foundation.application import Change, Metadata, Outcome, Publication
from operations.foundation.domain import record
from operations.foundation.identity import new_id
from operations.contexts.collection.application import CollectOrder
from operations.contexts.collection.domain import Pickup, PickupSnapshot
from operations.contexts.preparation.domain import PreparationTicket, TicketSnapshot


def present(row: Row | None) -> Row:
    assert row is not None
    return row


def test_atomic_receipts_realtime_rollback_concurrency_and_fencing() -> None:
    database = Database("preparation", os.environ["PREPARATION_DATABASE_URL"])
    def restore(value: object) -> TicketSnapshot:
        return PreparationTicket.restore(value).snapshot()
    commands, queries = Commands(database, "ticket", restore), Queries(database, "ticket", restore)
    identity = new_id()
    metadata = Metadata(new_id(), identity, "test.accept", new_id(), expected=0, input={"order": identity})
    state: TicketSnapshot = {"id": identity, "orderId": identity, "customerId": new_id(),
                            "instructions": "Coffee", "status": "queued"}
    try:
        # Encoding fails after the attempted state write: neither state nor receipts may survive.
        with pytest.raises(KeyError):
            commands.execute(metadata, lambda _: Change(state, "queued", publications=(Publication("invalid", {}),)))
        assert queries.get(identity) is None
        first = commands.execute(metadata, lambda _: Change(state, "queued"))
        assert commands.execute(metadata, lambda _: pytest.fail("Replay must not re-run a decision")) == first
        conflict = commands.execute(replace(metadata, input={"different": True}), lambda _: pytest.fail("Conflict"))
        assert conflict["rejection"]["code"] == "idempotency_conflict"
        with database.pool.connection() as connection:
            assert present(connection.execute("SELECT count(*) AS n FROM cafe.command_receipts WHERE aggregate_id=%s",
                                      (identity,)).fetchone())["n"] == 1
            assert present(connection.execute("SELECT count(*) AS n FROM cafe.realtime_publications WHERE aggregate_id=%s",
                                      (identity,)).fetchone())["n"] == 1
        def prepare(loaded: TicketSnapshot | None) -> Change[TicketSnapshot]:
            assert loaded is not None
            next_state: TicketSnapshot = {**loaded, "status": "preparing"}
            return Change(next_state, "preparing")
        def update(_: int) -> Outcome:
            return commands.execute(replace(metadata, id=new_id(), name="test.start", expected=1),
                                    prepare)
        with ThreadPoolExecutor(max_workers=2) as workers:
            outcomes = list(workers.map(update, range(2)))
        assert sum("rejection" not in item for item in outcomes) == 1
        loaded = queries.get(identity)
        assert loaded is not None and loaded["version"] == 2
        row = claim(database, True)
        assert row is not None
        # Claim recovery is independent of clocks in the test process.
        with database.pool.connection() as connection:
            connection.execute("UPDATE cafe.realtime_dispatches SET lease_until=clock_timestamp()-interval '1 second' WHERE event_id=%s",
                               (row["id"],))
        reclaimed = claim(database, True)
        assert reclaimed is not None
        assert reclaimed["id"] == row["id"]
        finish(database, row, True)
        with database.pool.connection() as connection:
            assert present(connection.execute("SELECT completed_at FROM cafe.realtime_dispatches WHERE event_id=%s",
                                      (row["id"],)).fetchone())["completed_at"] is None
        finish(database, reclaimed, True)
        with database.pool.connection() as connection:
            assert present(connection.execute("SELECT completed_at FROM cafe.realtime_dispatches WHERE event_id=%s",
                                      (row["id"],)).fetchone())["completed_at"] is not None
    finally:
        database.pool.close()


def test_corrupt_pickup_does_not_commit_a_command_or_consumer_rejection() -> None:
    database = Database("collection", os.environ["COLLECTION_DATABASE_URL"])
    try:
        # Deliberately bypass domain restoration only to seed and repair corrupt storage.
        fixtures = Commands(database, "pickup", record)
        commands = Commands[PickupSnapshot](database, "pickup", lambda value: Pickup.restore(value).snapshot())
        identity = new_id()
        state: dict[str, object] = {"id": identity, "orderId": new_id(), "customerId": new_id(),
                 "code": "broken", "status": "ready"}
        fixtures.execute(Metadata(new_id(), identity, "test.fixture", new_id(), expected=0),
                         lambda _: Change(state, "ready"))
        metadata = Metadata(new_id(), identity, "collection.CollectOrder", new_id(), input={"code": "ABC123"},
                            consumer="test.corrupt-pickup", source_id=new_id(), source_hash="fixture")
        with pytest.raises(ValueError):
            CollectOrder(commands).execute(metadata, {"code": "ABC123"})
        with database.pool.connection() as connection:
            assert present(connection.execute("SELECT count(*) AS n FROM cafe.command_receipts WHERE command_id=%s",
                                      (metadata.id,)).fetchone())["n"] == 0
            assert present(connection.execute("SELECT count(*) AS n FROM cafe.consumer_receipts WHERE event_id=%s",
                                      (metadata.source_id,)).fetchone())["n"] == 0
        repaired = {**state, "code": "ABC123"}
        fixtures.execute(Metadata(new_id(), identity, "test.repair", new_id(), expected=1),
                         lambda _: Change(repaired, "ready"))
        assert CollectOrder(commands).execute(metadata, {"code": "ABC123"})["status"] == "collected"
    finally:
        database.pool.close()
