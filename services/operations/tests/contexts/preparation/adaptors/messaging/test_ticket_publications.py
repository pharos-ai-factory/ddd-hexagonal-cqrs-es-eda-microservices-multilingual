"""Completing a ticket stages a decodable private fact for independent publication."""

from operations.adaptors.generated.message_names import PreparationDomainEvent

from operations.adaptors.command_execution import EventRecordingWriteRepository
from operations.adaptors.messaging.event_codec import decode, encode
from operations.contexts.preparation.adaptors.messaging.private_events import DRINKS_READY_DOMAIN
from operations.contexts.preparation.adaptors.messaging.ticket_publications import ticket_publications
from operations.contexts.preparation.application.commands.complete_preparation import (
    CompletePreparationCommandHandler,
)
from operations.contracts.requests.generated.preparation import CompletePreparationCommand
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketState
from operations.foundation.application import Loaded, Metadata
from operations.foundation.write_repository import CommandContext

TICKET = "11111111-1111-4111-8111-111111111111"
ORDER = "22222222-2222-4222-8222-222222222222"
CUSTOMER = "33333333-3333-4333-8333-333333333333"


class TicketRepositoryProbe:
    """Expose one preparing ticket and record its staged aggregate save."""

    def __init__(self) -> None:
        self.ticket = PreparationTicket(TicketState(TICKET, ORDER, CUSTOMER, "One coffee", "preparing"))
        self.saved = False

    def get(self, identity: str) -> Loaded[PreparationTicket]:
        assert identity == TICKET
        return Loaded(exists=True, version=2, state=self.ticket)

    def save(self, aggregate: PreparationTicket) -> None:
        assert aggregate is self.ticket
        self.saved = True


def test_complete_preparation_stages_private_drinks_ready_with_original_terms() -> None:
    source = TicketRepositoryProbe()
    repository = EventRecordingWriteRepository(source, ticket_publications)
    result = CompletePreparationCommandHandler(repository).execute(
        CommandContext(TICKET),
        CompletePreparationCommand(),
    )
    assert result.status == "ready" and source.saved
    assert source.ticket.snapshot()["status"] == "ready"
    assert len(repository.publications) == 1
    private = repository.publications[0]
    assert private.name == PreparationDomainEvent.DRINKS_READY
    metadata = Metadata(ORDER, TICKET, "complete", CUSTOMER)
    _, body = encode(
        "preparation", "ticket", TICKET, 3, metadata, private, occurred_at="2026-10-14T10:00:00Z"
    )
    wire, _ = decode(body)
    assert wire.visibility == "domain" and wire.aggregate_version == 3
    event = DRINKS_READY_DOMAIN.parse(wire)
    assert (event.order_id, event.customer_id) == (ORDER, CUSTOMER)
