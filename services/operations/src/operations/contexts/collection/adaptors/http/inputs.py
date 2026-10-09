"""Collection HTTP command input validation."""
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommand
from operations.foundation.domain import record, text


def collect_order(value: object) -> CollectOrderCommand:
    fields = record(value)
    if set(fields) != {"code"}:
        raise ValueError("The collection code is required")
    return CollectOrderCommand(code=text(fields["code"]))
