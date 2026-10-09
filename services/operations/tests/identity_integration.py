"""A valid UUID in stored state must still identify the locked aggregate."""
import os
import pytest
from psycopg import Connection, connect
from psycopg.rows import dict_row
from psycopg.types.json import Jsonb
from operations.adaptors.postgres import PostgresAggregateCommandStore, PostgresContextDatabase, PostgresAggregateQueries, Row
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommandHandler
from operations.contexts.collection.domain.pickup import Pickup, PickupSnapshot
from operations.foundation.application import Change, Metadata
from operations.foundation.domain import CorruptState
from operations.foundation.identity import new_id
from tests.persistence_integration import present


@pytest.mark.parametrize("matching_identity", [True, False], ids=["matching", "different_valid_uuid"])
def test_pickup_identity_must_match_its_storage_key(matching_identity: bool) -> None:
    assert os.environ["CAFE_DISPOSABLE_PROJECT"].startswith("cafe-reference-test-")
    database = PostgresContextDatabase("collection", os.environ["COLLECTION_DATABASE_URL"])
    identity, source = new_id(), new_id()
    state: PickupSnapshot = {"id": identity if matching_identity else new_id(),
        "orderId": new_id(), "customerId": new_id(), "code": "ABC123", "status": "ready"}
    metadata = Metadata(new_id(), identity, "collection.CollectOrder", new_id(), expected=1,
        input={"code": "ABC123"}, consumer="collection.test-identity", source_id=source, source_hash="fixture")
    admin: Connection[Row] = connect(os.environ["DATABASE_ADMIN_URL"], dbname="cafe_collection",
                                     row_factory=dict_row, autocommit=True)
    try:
        # Only the disposable administrator can introduce or repair invalid authority.
        with admin.transaction():
            admin.execute("SELECT set_config('cafe.command_target',%s,true)", ("pickup:"+identity,))
            admin.execute("INSERT INTO cafe.aggregates(kind,id,version,state) VALUES('pickup',%s,1,%s)",
                          (identity, Jsonb(state)))
        commands = PostgresAggregateCommandStore[PickupSnapshot](database, "pickup", lambda value: Pickup.restore(value).snapshot())
        handler = CollectOrderCommandHandler(commands)
        if not matching_identity:
            with pytest.raises(CorruptState):
                handler.execute(metadata, {"code": "ABC123"})
            stored = present(admin.execute("SELECT version,state FROM cafe.aggregates WHERE kind='pickup' AND id=%s",
                                           (identity,)).fetchone())
            assert stored == {"version": 1, "state": state}
            evidence(admin, identity, source, 0)
            with pytest.raises(CorruptState):
                PostgresAggregateQueries(database, "pickup", lambda value: Pickup.restore(value).snapshot()).get(identity)
            with admin.transaction():
                admin.execute("SET LOCAL session_replication_role='replica'")
                admin.execute("UPDATE cafe.aggregates SET state=%s WHERE kind='pickup' AND id=%s",
                              (Jsonb({**state, "id": identity}), identity))
        # Repair preserves both the expected version and the original attempt identity.
        outcome = handler.execute(metadata, {"code": "ABC123"})
        assert outcome == {"aggregateId": identity, "version": 2, "status": "collected"}
        assert handler.execute(metadata, {"code": "ABC123"}) == outcome
        stored = present(admin.execute("SELECT version,state FROM cafe.aggregates WHERE kind='pickup' AND id=%s",
                                       (identity,)).fetchone())
        assert stored == {"version": 2, "state": {**state, "id": identity, "status": "collected"}}
        evidence(admin, identity, source, 1)
    finally:
        admin.close()
        database.pool.close()


def test_proposed_pickup_identity_cannot_change_the_command_target() -> None:
    assert os.environ["CAFE_DISPOSABLE_PROJECT"].startswith("cafe-reference-test-")
    database = PostgresContextDatabase("collection", os.environ["COLLECTION_DATABASE_URL"])
    identity, source = new_id(), new_id()
    state: PickupSnapshot = {"id": new_id(), "orderId": new_id(), "customerId": new_id(),
                            "code": "ABC123", "status": "ready"}
    metadata = Metadata(new_id(), identity, "collection.test-proposal", new_id(), expected=0,
                        input={}, consumer="collection.test-proposal", source_id=source, source_hash="fixture")
    admin: Connection[Row] = connect(os.environ["DATABASE_ADMIN_URL"], dbname="cafe_collection",
                                     row_factory=dict_row, autocommit=True)
    try:
        commands = PostgresAggregateCommandStore[PickupSnapshot](database, "pickup", lambda value: Pickup.restore(value).snapshot())
        with pytest.raises(CorruptState):
            commands.execute(metadata, lambda _: Change(state, "ready", True))
        assert admin.execute("SELECT id FROM cafe.aggregates WHERE id=%s", (identity,)).fetchone() is None
        evidence(admin, identity, source, 0)
        state["id"] = identity
        assert commands.execute(metadata, lambda _: Change(state, "ready", True))["version"] == 1
    finally:
        admin.close()
        database.pool.close()


def evidence(admin: Connection[Row], identity: str, source: str, expected: int) -> None:
    for table in ("command_receipts", "outbox_events", "realtime_publications"):
        assert present(admin.execute(f"SELECT count(*) AS n FROM cafe.{table} WHERE aggregate_id=%s",
                                     (identity,)).fetchone())["n"] == expected
    assert present(admin.execute("SELECT count(*) AS n FROM cafe.consumer_receipts WHERE event_id=%s",
                                 (source,)).fetchone())["n"] == expected
