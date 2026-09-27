from dataclasses import dataclass, replace
from typing import ClassVar, Literal, TypedDict
from operations.foundation.domain import CorruptState, Rejection, identifier, record, text

type TicketStatus = Literal["queued", "preparing", "ready"]


class TicketSnapshot(TypedDict):
    id: str
    orderId: str
    customerId: str
    instructions: str
    status: TicketStatus


@dataclass(frozen=True)
class TicketState:
    id: str
    order_id: str
    customer_id: str
    instructions: str
    status: TicketStatus = "queued"


@dataclass(frozen=True)
class PreparationStarted:
    name: ClassVar[str] = "PreparationStarted"
    state: TicketState


@dataclass(frozen=True)
class DrinksReady:
    name: ClassVar[str] = "DrinksReady"
    order_id: str
    customer_id: str


type TicketFact = PreparationStarted | DrinksReady


class PreparationTicket:
    def __init__(self, state: TicketState) -> None:
        for value in (state.id, state.order_id, state.customer_id):
            identifier(value)
        if (not isinstance(state.instructions, str) or not state.instructions
            or state.status not in ("queued", "preparing", "ready")):
            raise ValueError("Invalid preparation ticket state")
        self._state = state
        self._facts: list[TicketFact] = []

    @classmethod
    def restore(cls, value: object) -> "PreparationTicket":
        try:
            fields = record(value)
            match fields["status"]:
                case "queued" | "preparing" | "ready" as status:
                    return cls(TicketState(identifier(fields["id"]), identifier(fields["orderId"]),
                        identifier(fields["customerId"]), text(fields["instructions"]), status))
                case _:
                    raise ValueError("Invalid preparation status")
        except (Rejection, KeyError, TypeError, ValueError) as error:
            raise CorruptState("Corrupt preparation ticket state") from error

    def start(self) -> None:
        if self._state.status != "queued":
            raise Rejection("ticket_not_queued", "Only queued tickets can start")
        self._state = replace(self._state, status="preparing")
        self._facts.append(PreparationStarted(self._state))

    def complete(self) -> None:
        if self._state.status != "preparing":
            raise Rejection("preparation_not_started", "Only preparing tickets can become ready")
        self._state = replace(self._state, status="ready")
        self._facts.append(DrinksReady(self._state.order_id, self._state.customer_id))

    def snapshot(self) -> TicketSnapshot:
        s = self._state
        return {"id": s.id, "orderId": s.order_id, "customerId": s.customer_id,
                "instructions": s.instructions, "status": s.status}

    def events(self) -> tuple[TicketFact, ...]:
        return tuple(self._facts)
