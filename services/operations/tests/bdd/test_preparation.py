import pytest
from pytest_bdd import given, when, then, parsers, scenarios

from operations.contexts.preparation.application.commands.accept_order import AcceptOrderCommand, AcceptOrderCommandHandler
from operations.contexts.preparation.application.event_handlers import OrderPlacedIntegrationEventHandler
from operations.contexts.preparation.application.commands.start_preparation import StartPreparationCommandHandler
from operations.contexts.preparation.application.commands.complete_preparation import CompletePreparationCommandHandler
from operations.contexts.preparation.domain import TicketSnapshot
from operations.contracts.events import OrderPlaced
from operations.foundation.application import Metadata, Outcome
from tests.bdd.conftest import SPECIFICATIONS, CUSTOMER, ORDER, ROOT, CommandProbe

type PreparationProbe = CommandProbe[TicketSnapshot, OrderPlaced]

scenarios(str(SPECIFICATIONS / "preparation"))


@pytest.fixture
def probe() -> PreparationProbe:
    return CommandProbe()


@given(parsers.parse('an order for {quantity:d} "{name}" drinks has been placed'))
def placed(probe: PreparationProbe, quantity: int, name: str) -> None:
    probe.event = {"orderId": ORDER, "customerId": CUSTOMER, "editionId": ROOT, "currency": "EUR",
        "lines": [{"id": ROOT, "offerCode": "C1", "quantity": quantity, "name": name, "minor": 300}]}


@when("Preparation accepts the placed order")
def accept(probe: PreparationProbe, metadata: Metadata) -> None:
    commands: list[AcceptOrderCommand] = []
    class QueueProbe:
        def enqueue(self, incoming: Metadata, command: AcceptOrderCommand) -> Outcome:
            commands.append(command)
            return Outcome(aggregateId=incoming.target, version=0, status="queued")
    OrderPlacedIntegrationEventHandler(QueueProbe()).handle(metadata, probe.incoming())
    AcceptOrderCommandHandler(probe).execute(metadata, commands[0])


@given("Preparation has accepted the placed order")
def accepted(probe: PreparationProbe, metadata: Metadata) -> None:
    accept(probe, metadata)
    probe.succeeded()


@when("the barista starts the ticket")
def start(probe: PreparationProbe, metadata: Metadata) -> None:
    StartPreparationCommandHandler(probe).execute(metadata, {})


@given("the ticket is being prepared")
def preparing(probe: PreparationProbe, metadata: Metadata) -> None:
    accepted(probe, metadata)
    start(probe, metadata)
    probe.succeeded()


@when("the barista completes the ticket")
def complete(probe: PreparationProbe, metadata: Metadata) -> None:
    CompletePreparationCommandHandler(probe).execute(metadata, {})


@given("the ticket has been completed")
def completed(probe: PreparationProbe, metadata: Metadata) -> None:
    preparing(probe, metadata)
    complete(probe, metadata)
    probe.succeeded()


@then(parsers.parse('the ticket is "{status}" with instructions "{instructions}"'))
def ticket(probe: PreparationProbe, status: str, instructions: str) -> None:
    probe.succeeded()
    assert probe.current()["status"] == status
    assert probe.current()["instructions"] == instructions
    assert probe.current()["orderId"] == ORDER
    assert probe.current()["customerId"] == CUSTOMER


@then("Preparation produces no outgoing events")
def no_events(probe: PreparationProbe) -> None:
    assert probe.publications == ()


@then(parsers.parse('Preparation rejects the command with "{code}"'))
def rejection(probe: PreparationProbe, code: str) -> None:
    assert probe.rejection and probe.rejection["code"] == code


@then("the ticket and its outgoing events are unchanged")
def unchanged(probe: PreparationProbe) -> None:
    probe.unchanged()


@then("one DrinksReady publication identifies the order and customer")
def ready_event(probe: PreparationProbe) -> None:
    assert len(probe.publications) == 1
    event = probe.publications[0]
    assert event.name == "preparation.drinks-ready"
    assert event.payload == {"orderId": ORDER, "customerId": CUSTOMER}
