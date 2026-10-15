from operations.contracts.events.generated import ordering as events
from operations.foundation.application import CommandBusPort, Metadata, Outcome
from operations.contracts.requests.generated.preparation import AcceptOrderCommand


class OrderPlacedIntegrationEventHandler:
    """Translate a published order into a durable Preparation command."""

    def __init__(self, commands: CommandBusPort[AcceptOrderCommand]) -> None:
        self.commands = commands

    def handle(self, metadata: Metadata, event: events.OrderPlacedIntegrationEvent) -> Outcome:
        return self.commands.enqueue(
            metadata,
            AcceptOrderCommand(
                orderId=event["orderId"],
                customerId=event["customerId"],
                instructions="; ".join(f"{line['quantity']} × {line['name']}" for line in event["lines"]),
            ),
        )
