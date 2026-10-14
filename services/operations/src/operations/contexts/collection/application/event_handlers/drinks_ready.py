from operations.contracts.events.generated import preparation as events
from operations.foundation.application import DurableCommandPort, Metadata, Outcome
from operations.contexts.collection.application.commands.open_pickup import OpenPickupCommand


class DrinksReadyIntegrationEventHandler:
    """Translate Preparation's published fact into Collection's durable command."""

    def __init__(self, commands: DurableCommandPort[OpenPickupCommand]) -> None:
        self.commands = commands

    def handle(self, metadata: Metadata, event: events.DrinksReadyIntegrationEvent) -> Outcome:
        return self.commands.enqueue(
            metadata, OpenPickupCommand(orderId=event["orderId"], customerId=event["customerId"])
        )
