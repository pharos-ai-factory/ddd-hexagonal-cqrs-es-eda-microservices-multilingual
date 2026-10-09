from operations.contexts.collection.adaptors.messaging.pickup_publications import pickup_publications
from operations.adaptors.command_execution import CommandExecutor
from operations.contexts.collection.domain.pickup import Pickup
import re
import pytest

from pytest_bdd import given, when, then, parsers, scenarios

from operations.contexts.collection.application.commands.open_pickup import OpenPickupCommand, OpenPickupCommandHandler
from operations.contexts.collection.application.event_handlers.drinks_ready import DrinksReadyIntegrationEventHandler
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommandHandler
from operations.contexts.collection.domain.pickup import PickupSnapshot
from operations.contracts.events import DrinksReady
from operations.foundation.application import Metadata, Outcome
from operations.foundation.identity import derived_id
from tests.bdd.conftest import SPECIFICATIONS, CUSTOMER, ORDER, ROOT, CommandProbe

type CollectionProbe = CommandProbe[PickupSnapshot, DrinksReady]

scenarios(str(SPECIFICATIONS / "collection"))


@pytest.fixture
def probe() -> CollectionProbe:
    return CommandProbe()


@given("drinks are ready for an order")
def ready(probe: CollectionProbe) -> None:
    probe.event = {"orderId": ORDER, "customerId": CUSTOMER}


@when("Collection handles the ready drinks")
def open_pickup(probe: CollectionProbe, metadata: Metadata) -> None:
    commands: list[OpenPickupCommand] = []
    class QueueProbe:
        def enqueue(self, incoming: Metadata, command: OpenPickupCommand) -> Outcome:
            commands.append(command)
            return Outcome(aggregateId=incoming.target, version=0, status="queued")
    DrinksReadyIntegrationEventHandler(QueueProbe()).handle(metadata, probe.incoming())
    CommandExecutor(probe.transaction(Pickup.restore, Pickup.snapshot, pickup_publications), lambda repository: OpenPickupCommandHandler(repository, derived_id)).execute(metadata, commands[0])


@given("the pickup has been opened")
def opened(probe: CollectionProbe, metadata: Metadata) -> None:
    open_pickup(probe, metadata)
    probe.succeeded()


@when("the customer presents the correct collection code")
def collect(probe: CollectionProbe, metadata: Metadata) -> None:
    CommandExecutor(probe.transaction(Pickup.restore, Pickup.snapshot, pickup_publications), lambda repository: CollectOrderCommandHandler(repository)).execute(metadata, {"code": probe.current()["code"]})


@given("the pickup has been collected")
def collected(probe: CollectionProbe, metadata: Metadata) -> None:
    opened(probe, metadata)
    collect(probe, metadata)
    probe.succeeded()


@when("the customer presents the wrong collection code")
def wrong_code(probe: CollectionProbe, metadata: Metadata) -> None:
    assert probe.current()["code"] != "WRONG1"
    CommandExecutor(probe.transaction(Pickup.restore, Pickup.snapshot, pickup_publications), lambda repository: CollectOrderCommandHandler(repository)).execute(metadata, {"code": "WRONG1"})


@then("the pickup is ready with a six-character collection code")
def pickup(probe: CollectionProbe) -> None:
    probe.succeeded()
    assert probe.current()["status"] == "ready"
    assert re.fullmatch(r"[A-Z0-9]{6}", probe.current()["code"])
    assert probe.current()["code"] == derived_id("collection-code", ORDER)[:6].upper()


@then("one PickupOpened publication contains that code, order and customer")
def opened_event(probe: CollectionProbe) -> None:
    assert len(probe.publications) == 1
    event = probe.publications[0]
    assert event.name == "collection.pickup-opened"
    assert event.payload == {
        "pickupId": ROOT, "orderId": ORDER, "customerId": CUSTOMER, "collectionCode": probe.current()["code"]}


@then(parsers.parse('Collection rejects the command with "{code}"'))
def rejection(probe: CollectionProbe, code: str) -> None:
    assert probe.rejection and probe.rejection["code"] == code


@then("the pickup and its outgoing events are unchanged")
def unchanged(probe: CollectionProbe) -> None:
    probe.unchanged()


@then("the pickup is collected")
def collected_state(probe: CollectionProbe) -> None:
    probe.succeeded()
    assert probe.current()["status"] == "collected"


@then("one OrderCollected publication identifies the order and customer")
def collected_event(probe: CollectionProbe) -> None:
    assert len(probe.publications) == 1
    event = probe.publications[0]
    assert event.name == "collection.order-collected"
    assert event.payload == {"orderId": ORDER, "customerId": CUSTOMER}
