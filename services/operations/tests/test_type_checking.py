"""Exercise the configured gate with valid callers and concrete contract mistakes."""

from pathlib import Path
from mypy import api

IMPORTS = """\
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommandHandler
from operations.contexts.collection.domain.pickup import Pickup, PickupSnapshot
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket
from operations.contracts.events.generated.preparation import DrinksReadyIntegrationEvent
from operations.foundation.write_repository import CommandContext, WriteRepository
from operations.foundation.application import RealtimeBusPort, RealtimeMessage

def caller(pickups: WriteRepository[Pickup], tickets: WriteRepository[PreparationTicket], metadata: CommandContext, realtime: RealtimeBusPort[RealtimeMessage]) -> None:
    collect = CollectOrderCommandHandler(pickups, realtime)
    collect.execute(metadata, {"code": "ABC123"})
    ready: DrinksReadyIntegrationEvent = {"orderId": metadata.target, "customerId": metadata.target}
"""


def test_type_gate_rejects_wrong_commands_events_states_and_ports(tmp_path: Path) -> None:
    source = tmp_path / "caller.py"
    config = Path(__file__).resolve().parents[1] / "pyproject.toml"
    arguments = ["--config-file", str(config), "--follow-imports=silent", str(source)]
    source.write_text(IMPORTS)
    output, errors, status = api.run(arguments)
    assert status == 0, output + errors

    mistakes = [
        '    collect.execute(metadata, {"code": 123})',
        "    collect.execute(metadata, {})",
        "    CollectOrderCommandHandler(tickets, realtime)",
        "    CollectOrderCommandHandler(pickups, pickups)",
        '    missing_order: DrinksReadyIntegrationEvent = {"customerId": metadata.target}',
        '    wrong_status: PickupSnapshot = {"id": metadata.target, "orderId": metadata.target, '
        '"customerId": metadata.target, "code": "ABC123", "status": "queued"}',
    ]
    source.write_text(IMPORTS + "\n".join(mistakes) + "\n")
    output, errors, status = api.run(arguments)
    assert status == 1, output + errors
    for line in range(len(IMPORTS.splitlines()) + 1, len(IMPORTS.splitlines()) + len(mistakes) + 1):
        assert f"caller.py:{line}: error:" in output, output


def test_registration_identifiers_are_typed(tmp_path: Path) -> None:
    source = tmp_path / "registration.py"
    config = Path(__file__).resolve().parents[1] / "pyproject.toml"
    imports = """from operations.adaptors.generated.request_names import CommandName, QueryName, PreparationCommand, PreparationQuery
from operations.adaptors.generated.message_names import EventName, PreparationDomainEvent, SubscriptionName, PreparationSubscription
command: CommandName = PreparationCommand.START_PREPARATION
query: QueryName = PreparationQuery.GET_TICKET
event: EventName = PreparationDomainEvent.DRINKS_READY
subscription: SubscriptionName = PreparationSubscription.ACCEPT_ORDER
"""
    arguments = ["--config-file", str(config), "--follow-imports=silent", str(source)]
    source.write_text(imports)
    output, errors, status = api.run(arguments)
    assert status == 0, output + errors
    mistakes = [
        'command = "startPreparation"',
        "query = PreparationCommand.START_PREPARATION",
        'event = "preparation.drinks-ready.domain"',
        "subscription = PreparationDomainEvent.DRINKS_READY",
    ]
    source.write_text(imports + "\n".join(mistakes) + "\n")
    output, errors, status = api.run(arguments)
    assert status == 1, output + errors
    for line in range(len(imports.splitlines()) + 1, len(imports.splitlines()) + len(mistakes) + 1):
        assert f"registration.py:{line}: error:" in output, output


def test_generated_query_binding_rejects_another_response(tmp_path: Path) -> None:
    source = tmp_path / "responses.py"
    config = Path(__file__).resolve().parents[1] / "pyproject.toml"
    source.write_text("""from operations.adaptors.messaging.rabbitmq_request_registry import RabbitMQRequestRegistry
from operations.contexts.preparation.adaptors.messaging.generated.response_payloads import register_get_ticket
from operations.contexts.preparation.adaptors.messaging.ticket_requests import get_ticket
from operations.contracts.requests.generated.preparation import GetTicketQuery
from operations.contracts.requests.generated.preparation_responses import GetTicketResponse, ListTicketsResponse

def handler(query: GetTicketQuery) -> GetTicketResponse:
    return {"error": {"code": "not_found", "message": ""}}
register_get_ticket(RabbitMQRequestRegistry("preparation"), get_ticket, handler)
""")
    args = ["--config-file", str(config), "--follow-imports=silent", str(source)]
    output, errors, status = api.run(args)
    assert status == 0, output + errors
    source.write_text(source.read_text().replace("-> GetTicketResponse:", "-> ListTicketsResponse:"))
    output, errors, status = api.run(args)
    assert status == 1 and "register_get_ticket" in output and "incompatible type" in output, output + errors
