"""Validate persisted ticket state for both command and query adaptors."""
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketSnapshot


def restore_ticket_snapshot(value: object) -> TicketSnapshot:
    """Rehydrate the owner aggregate and return its validated storage snapshot."""
    return PreparationTicket.restore(value).snapshot()
