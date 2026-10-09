"""Preparation HTTP command input validation."""
from operations.contexts.preparation.application.commands.start_preparation import StartPreparationCommand
from operations.contexts.preparation.application.commands.complete_preparation import CompletePreparationCommand
from operations.foundation.domain import record


def empty(value: object) -> None:
    if record(value):
        raise ValueError("This command requires an empty object")


def start_preparation(value: object) -> StartPreparationCommand:
    empty(value)
    return StartPreparationCommand()


def complete_preparation(value: object) -> CompletePreparationCommand:
    empty(value)
    return CompletePreparationCommand()
