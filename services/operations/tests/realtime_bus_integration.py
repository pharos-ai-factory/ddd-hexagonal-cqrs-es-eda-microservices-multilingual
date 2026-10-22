"""Real transactions prove opt-in publication, rollback, receipts and application outcomes."""

import os
from dataclasses import replace
import pytest
from operations.adaptors.command_execution import TransactionResult
from operations.adaptors.persistence.aggregate_transaction import PostgresAggregateTransaction
from operations.adaptors.persistence.postgres_database import PostgresContextDatabase
from operations.adaptors.persistence.realtime_event_bus import PostgresRealtimeMessageBus
from operations.adaptors.generated.cafe.realtime.v1.realtime_pb2 import Publication
from operations.contexts.preparation.adaptors.persistence.ticket_snapshot import restore_ticket_snapshot
from operations.contexts.preparation.domain.preparation_ticket import TicketSnapshot
from operations.foundation.application import (
    Metadata,
    RealtimeMessage,
    context_target,
    principal_target,
    rejection_message,
    snapshot_message,
)
from operations.foundation.identity import new_id
from operations.foundation.domain import integer, RejectionDetail
from operations.adaptors.persistence.dispatch_leases import binary
from operations.foundation.write_repository import WriteRepository


def test_explicit_realtime_and_subscription_recovery() -> None:
    db = PostgresContextDatabase("preparation", os.environ["PREPARATION_DATABASE_URL"])
    target = new_id()
    state: TicketSnapshot = {
        "id": target,
        "orderId": new_id(),
        "customerId": new_id(),
        "instructions": "Tea",
        "status": "queued",
    }
    unit = PostgresAggregateTransaction[TicketSnapshot, TicketSnapshot](
        db, "ticket", restore_ticket_snapshot, lambda r: r
    )
    command = Metadata(new_id(), target, "test.explicit", new_id(), expected=0)

    def count() -> int:
        with db.pool.connection() as connection:
            row = connection.execute(
                "SELECT count(*) AS n FROM cafe.realtime_publications WHERE aggregate_id=%s", (target,)
            ).fetchone()
            assert row
            return integer(row["n"])

    def save(repository: WriteRepository[TicketSnapshot]) -> TransactionResult:
        repository.save(state)
        return TransactionResult("queued")

    try:
        assert unit.execute(command, save)["version"] == 1
        assert count() == 0
        ready: TicketSnapshot = {**state, "status": "ready"}

        def invalid(repository: WriteRepository[TicketSnapshot]) -> TransactionResult:
            repository.save(ready)
            return TransactionResult(
                "ready", realtime=(snapshot_message(context_target("preparation"), state),)
            )

        changed = replace(command, id=new_id(), expected=1)
        with pytest.raises(ValueError, match="differs from committed"):
            unit.execute(changed, invalid)

        def explicit(repository: WriteRepository[TicketSnapshot]) -> TransactionResult:
            repository.save(ready)
            return TransactionResult(
                "ready", realtime=(snapshot_message(context_target("preparation"), ready),)
            )

        assert unit.execute(changed, explicit)["version"] == 2
        unit.execute(changed, lambda _: pytest.fail("Duplicate command executed"))
        assert count() == 1
        source = Metadata(
            new_id(),
            target,
            "realtime",
            command.correlation,
            consumer="preparation.realtime-test",
            source_id=new_id(),
            source_hash="a" * 64,
            source_version=1,
        )
        bus = PostgresRealtimeMessageBus(db, source.consumer, "ticket")
        event = RealtimeMessage(
            target=context_target("preparation"), source=source, revision=2, snapshot=ready
        )
        bus.publish(event)
        bus.publish(event)
        assert count() == 2
        with pytest.raises(ValueError, match="Conflicting"):
            bus.publish(replace(event, source=replace(source, source_hash="b" * 64)))
        with db.pool.connection() as connection:
            rows = connection.execute(
                "SELECT body FROM cafe.realtime_publications WHERE aggregate_id=%s", (target,)
            ).fetchall()
            assert all(Publication.FromString(binary(row["body"])).revision == 2 for row in rows)

        missing = replace(command, id=new_id(), target=new_id(), actor_id=new_id())
        rejection: RejectionDetail = {"code": "not_found", "message": "The ticket does not exist"}
        outcome = unit.execute(
            missing,
            lambda _: TransactionResult(
                "",
                realtime=(rejection_message(principal_target(missing.actor_id), rejection),),
                rejection=rejection,
            ),
        )
        assert outcome["version"] == 0 and outcome["rejection"] == rejection
        with db.pool.connection() as connection:
            row = connection.execute(
                "SELECT body FROM cafe.realtime_publications WHERE aggregate_id=%s", (missing.target,)
            ).fetchone()
            assert row
            wire = Publication.FromString(binary(row["body"]))
            assert wire.operation_rejected.command_id == missing.id
    finally:
        db.pool.close()
