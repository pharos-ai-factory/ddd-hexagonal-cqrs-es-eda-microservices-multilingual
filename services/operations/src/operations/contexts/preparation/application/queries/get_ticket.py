from operations.contracts.requests.generated.preparation import GetTicketQuery
from operations.contracts.requests.generated.preparation_responses import GetTicketResponse, LoadedTicket
from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository


class GetTicketQueryHandler:
    """Preserve revision, missing-resource and pagination policy in a generated response."""

    def __init__(self, read_repository: TicketReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: GetTicketQuery) -> GetTicketResponse:
        value = self.read_repository.get(query.id)
        if value is None:
            return {"error": {"code": "not_found", "message": ""}}
        return {
            "ticket": LoadedTicket(exists=value["exists"], version=value["version"], state=value["state"])
        }
