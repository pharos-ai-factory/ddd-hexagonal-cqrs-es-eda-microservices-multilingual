"""Authoritative aggregate persistence inside the preparation transaction boundary."""
from operations.foundation.write_repository import WriteRepository
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket

type TicketWriteRepository = WriteRepository[PreparationTicket]
