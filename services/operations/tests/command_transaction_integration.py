"""Prove local transaction scope, staged saves and repository lifetime on PostgreSQL."""
from dataclasses import replace
import os
import pytest
from operations.adaptors.postgres import PostgresContextDatabase
from operations.adaptors.snapshot_read_repository import PostgresSnapshotReadRepository
from operations.adaptors.aggregate_transaction import PostgresAggregateTransaction
from operations.contexts.preparation.adaptors.persistence.ticket_snapshot import restore_ticket_snapshot
from operations.contexts.preparation.domain.preparation_ticket import TicketSnapshot
from operations.foundation.application import Metadata
from operations.foundation.domain import CorruptState, Rejection
from operations.foundation.identity import new_id
from operations.foundation.write_repository import WriteRepository
from operations.adaptors.command_execution import TransactionResult as CommandResult


def test_local_transaction_scope_lifetime_and_rejected_save() -> None:
    database = PostgresContextDatabase("preparation", os.environ["PREPARATION_DATABASE_URL"])
    unit = PostgresAggregateTransaction[TicketSnapshot, TicketSnapshot](database, "ticket", restore_ticket_snapshot, lambda repo: repo)
    queries = PostgresSnapshotReadRepository(database, "ticket", restore_ticket_snapshot)
    identity = new_id()
    state: TicketSnapshot = {"id": identity, "orderId": new_id(), "customerId": new_id(),
                             "instructions": "Coffee", "status": "queued"}
    metadata = Metadata(new_id(), identity, "test.scope", new_id(), expected=0)
    escaped: list[WriteRepository[TicketSnapshot]] = []
    try:
        def create(repository: WriteRepository[TicketSnapshot]) -> CommandResult:
            escaped.append(repository)
            assert repository.get(identity) is None
            with pytest.raises(ValueError, match="another aggregate"):
                repository.get(new_id())
            with pytest.raises(CorruptState):
                repository.save({**state, "id": new_id()})
            repository.save(state)
            # Saving captures the snapshot; later in-memory changes require another save.
            state["instructions"] = "Unsaved change"
            return CommandResult("queued")
        assert unit.execute(metadata, create)["version"] == 1
        with pytest.raises(RuntimeError, match="outside its command transaction"):
            escaped[-1].get(identity)
        with pytest.raises(RuntimeError, match="outside its command transaction"):
            escaped[-1].save(state)
        loaded = queries.get(identity)
        assert loaded and loaded["state"]["instructions"] == "Coffee"
        rejection_metadata = replace(metadata, id=new_id(), expected=1)
        def reject(repository: WriteRepository[TicketSnapshot]) -> CommandResult:
            escaped.append(repository)
            repository.save({**state, "status": "preparing"})
            raise Rejection("fixture_rejection", "Reject after staging")
        rejected = unit.execute(rejection_metadata, reject)
        assert rejected["rejection"]["code"] == "fixture_rejection"
        assert rejected["version"] == 1
        assert queries.get(identity) == loaded
        assert unit.execute(rejection_metadata, lambda _: pytest.fail("Rejected replay ran again")) == rejected
        with pytest.raises(RuntimeError):
            escaped[-1].save(state)
        def fail(repository: WriteRepository[TicketSnapshot]) -> CommandResult:
            escaped.append(repository)
            repository.save(state)
            raise RuntimeError("Connection fixture failure")
        with pytest.raises(RuntimeError, match="Connection fixture failure"):
            unit.execute(replace(metadata, id=new_id(), expected=1), fail)
        with pytest.raises(RuntimeError, match="outside its command transaction"):
            escaped[-1].get(identity)
        assert queries.get(identity) == loaded
        with database.pool.connection() as connection:
            receipt = connection.execute("SELECT count(*) AS n FROM cafe.command_receipts WHERE aggregate_id=%s", (identity,)).fetchone()
            realtime = connection.execute("SELECT count(*) AS n FROM cafe.realtime_publications WHERE aggregate_id=%s", (identity,)).fetchone()
            assert receipt and receipt["n"] == 2
            assert realtime and realtime["n"] == 1
    finally:
        database.pool.close()
