"""Translate transport query arguments into named Preparation use cases."""
from operations.foundation.application import Loaded
from operations.foundation.pagination import Page, PageRequest
from operations.contexts.preparation.application.read_models.ticket import TicketView
from operations.contexts.preparation.application.queries.get_ticket import GetTicketQuery, GetTicketQueryHandler
from operations.contexts.preparation.application.queries.list_tickets import ListTicketsQuery, ListTicketsQueryHandler


class TicketQueryEndpoints:
    """Translate RabbitMQ query arguments into named application use cases."""
    def __init__(self, get: GetTicketQueryHandler, listing: ListTicketsQueryHandler) -> None:
        self.get_handler, self.list_handler = get, listing

    def get(self, identity: str) -> Loaded[TicketView] | None:
        return self.get_handler.execute(GetTicketQuery(identity))

    def list(self) -> list[Loaded[TicketView]]:
        return self.list_handler.execute(ListTicketsQuery()).items

    def page(self, request: PageRequest) -> Page[TicketView]:
        return self.list_handler.execute(ListTicketsQuery(request))
