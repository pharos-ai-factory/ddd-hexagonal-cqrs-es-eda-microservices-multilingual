"""Exercise the configured gate with valid callers and concrete contract mistakes."""
from pathlib import Path
from mypy import api

IMPORTS = """\
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommandHandler
from operations.contexts.collection.domain import PickupSnapshot
from operations.contexts.preparation.domain import TicketSnapshot
from operations.contracts.events import DrinksReady
from operations.foundation.application import AggregateCommandPort, Metadata

def caller(pickups: AggregateCommandPort[PickupSnapshot], tickets: AggregateCommandPort[TicketSnapshot], metadata: Metadata) -> None:
    collect = CollectOrderCommandHandler(pickups)
    collect.execute(metadata, {"code": "ABC123"})
    ready: DrinksReady = {"orderId": metadata.target, "customerId": metadata.target}
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
        '    collect.execute(metadata, {})',
        '    CollectOrderCommandHandler(tickets)',
        '    missing_order: DrinksReady = {"customerId": metadata.target}',
        '    wrong_status: PickupSnapshot = {"id": metadata.target, "orderId": metadata.target, '
        '"customerId": metadata.target, "code": "ABC123", "status": "queued"}',
    ]
    source.write_text(IMPORTS + "\n".join(mistakes) + "\n")
    output, errors, status = api.run(arguments)
    assert status == 1, output + errors
    for line in range(len(IMPORTS.splitlines()) + 1, len(IMPORTS.splitlines()) + len(mistakes) + 1):
        assert f"caller.py:{line}: error:" in output, output
