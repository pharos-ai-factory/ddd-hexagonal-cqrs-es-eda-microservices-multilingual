"""Exercise published and durable command registrations through the production composition."""

from collections.abc import Callable
from threading import Event
from unittest.mock import Mock, patch

from dependency_injector import providers
import pytest

from operations.adaptors.command_execution import TransactionResult
from operations.adaptors.messaging.rabbitmq_consumer import EventSubscription
from operations.apps.composition.collection import CollectionContainer, requests
from operations.apps.composition.preparation import PreparationContainer, subscriptions
from operations.contexts.collection.domain.pickup import Pickup, PickupSnapshot
from operations.contexts.preparation.application.commands.accept_order import AcceptOrderCommand
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketSnapshot
from operations.contexts.preparation.adaptors.messaging.ticket_publications import ticket_publications
from operations.contexts.collection.adaptors.messaging.pickup_publications import pickup_publications
from operations.foundation.application import Metadata, Outcome
from operations.foundation.identity import derived_id
from operations.foundation.write_repository import WriteRepository
from tests.bdd.conftest import CommandProbe
from tests.test_requests import ID, command


def test_published_registration_runs_the_concrete_handler_in_its_transaction() -> None:
    probe: CommandProbe[PickupSnapshot, object] = CommandProbe(
        state={"id": ID, "orderId": ID, "customerId": ID, "code": "ABC123", "status": "ready"}
    )
    transaction = probe.transaction(Pickup.restore, Pickup.snapshot, pickup_publications)
    captured: list[Metadata] = []

    class Transaction:
        def execute(
            self, metadata: Metadata, work: Callable[[WriteRepository[Pickup]], TransactionResult]
        ) -> Outcome:
            captured.append(metadata)
            return transaction.execute(metadata, work)

    container = CollectionContainer()
    container.pickup_transaction.override(providers.Object(Transaction()))
    container.queries.override(providers.Object(Mock()))
    registry = requests(container)
    request = command()
    request.command.collect_order.code = "ABC123"
    reply = registry.handle(request)
    assert reply.outcome.status == "collected"
    assert probe.state is not None and probe.state["status"] == "collected"
    assert captured[0].id == ID and captured[0].expected == 0
    assert captured[0].name == "collection.CollectOrder"
    assert captured[0].input == {"code": "ABC123"}
    assert [fact.name for fact in probe.publications] == ["collection.order-collected.domain"]
    with pytest.raises(ValueError, match="Duplicate request handler"):
        container.commands().request(registry, "collectOrder", lambda _: {}, lambda _: Mock())


def test_private_registration_uses_the_declared_codec_and_owner_handler() -> None:
    probe: CommandProbe[TicketSnapshot, object] = CommandProbe()
    container = PreparationContainer()
    container.ticket_transaction.override(
        providers.Object(
            probe.transaction(PreparationTicket.restore, PreparationTicket.snapshot, ticket_publications)
        )
    )
    container.accepted.override(providers.Object(Mock()))
    container.publish_drinks_ready.override(providers.Object(Mock()))
    worker = next(worker for worker in subscriptions(container) if worker.command)
    value: AcceptOrderCommand = {"orderId": ID, "customerId": ID, "instructions": "No sugar"}
    metadata = Metadata(
        id=derived_id(worker.consumer, ID),
        target=ID,
        name=worker.consumer,
        correlation=ID,
        input=value,
        consumer=worker.consumer,
        source_id=ID,
        source_hash="a" * 64,
    )
    body = container.codec().encode(metadata, value)

    def receive(_url: str, subscription: EventSubscription[AcceptOrderCommand], _stop: Event) -> None:
        assert subscription.command_decoder is not None
        assert subscription.consumer == "preparation.accept-order"
        received, decoded = subscription.command_decoder(body)
        assert decoded == value
        assert received.source_hash == metadata.source_hash
        outcome = subscription.handle(received, decoded)
        assert outcome["status"] == "queued"

    with patch("operations.apps.runtime.consume", receive):
        worker.run("unused", Event())
    assert probe.state is not None
    assert probe.state["instructions"] == "No sugar"
    assert probe.state["orderId"] == ID
