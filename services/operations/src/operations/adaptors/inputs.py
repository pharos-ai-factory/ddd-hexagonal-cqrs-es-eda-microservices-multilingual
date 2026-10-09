"""HTTP command decoders; no framework or wire values reach application handlers."""
from operations.contexts.collection.application.commands.collect_order import CollectOrderCommand
from operations.contexts.preparation.application.commands.complete_preparation import CompletePreparationCommand
from operations.contexts.preparation.application.commands.start_preparation import StartPreparationCommand
from operations.foundation.domain import record, text


def empty(value: object) -> None:
    if record(value):
        raise ValueError("This command requires an empty object")


def start_preparation(value: object) -> StartPreparationCommand:
    empty(value)
    return StartPreparationCommand()


def complete_preparation(value: object) -> CompletePreparationCommand:
    empty(value)
    return CompletePreparationCommand()


def collect_order(value: object) -> CollectOrderCommand:
    fields = record(value)
    if set(fields) != {"code"}:
        raise ValueError("The collection code is required")
    return CollectOrderCommand(code=text(fields["code"]))
