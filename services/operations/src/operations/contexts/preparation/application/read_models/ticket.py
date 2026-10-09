from typing import Literal, TypedDict


class TicketView(TypedDict):
    """Application read contract for one ticket, independent of storage and transport."""
    id: str
    orderId: str
    customerId: str
    instructions: str
    status: Literal["queued", "preparing", "ready"]
