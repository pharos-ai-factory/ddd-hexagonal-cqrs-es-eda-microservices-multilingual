from operations.contracts import events
from operations.foundation.application import DurableCommandPort, Metadata, Outcome
from operations.contexts.preparation.application.commands.accept_order import AcceptOrderCommand


class OrderPlacedIntegrationEventHandler:
    """Translate a published order into a durable Preparation command."""
    def __init__(self, commands: DurableCommandPort[AcceptOrderCommand]) -> None:
        self.commands = commands

    def handle(self, metadata: Metadata, event: events.OrderPlaced) -> Outcome:
        return self.commands.enqueue(metadata, AcceptOrderCommand(orderId=event["orderId"],
            customerId=event["customerId"],
            instructions="; ".join(f'{line["quantity"]} × {line["name"]}' for line in event["lines"])))
