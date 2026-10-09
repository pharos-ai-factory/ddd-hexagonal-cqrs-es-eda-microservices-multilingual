from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketState
import pytest
from operations.foundation.domain import Rejection

IDS = [f"00000000-0000-4000-8000-{number:012d}" for number in range(1, 5)]

def test_preparation_rejects_skipped_step_without_mutation() -> None:
    ticket = PreparationTicket(TicketState(IDS[0], IDS[1], IDS[2], "2 × coffee"))
    before = ticket.snapshot()
    with pytest.raises(Rejection, match="Only preparing"):
        ticket.complete()
    assert ticket.snapshot() == before
    assert ticket.events() == ()
    ticket.start()
    ticket.complete()
    assert ticket.snapshot()["status"] == "ready"
    assert [fact.name for fact in ticket.events()] == ["PreparationStarted", "DrinksReady"]
    with pytest.raises(Rejection):
        ticket.complete()


@pytest.mark.parametrize("state", [{"id": IDS[0], "orderId": "broken", "customerId": IDS[1], "instructions": "Coffee", "status": "queued"},
 {"id": IDS[0], "orderId": IDS[1], "customerId": IDS[2], "instructions": 1, "status": "queued"}, {}])
def test_corrupt_restoration_is_retryable(state: object) -> None:
    with pytest.raises(ValueError):
        PreparationTicket.restore(state)
