from typing import Protocol
from operations.foundation.pagination import PagedQueryPort
from operations.contexts.preparation.application.read_models.ticket import TicketView


class TicketReadRepository(PagedQueryPort[TicketView], Protocol):
    """Read tickets with stable revisions and optional keyset pagination."""
