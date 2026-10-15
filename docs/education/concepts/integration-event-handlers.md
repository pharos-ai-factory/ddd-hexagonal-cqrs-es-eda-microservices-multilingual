# Integration event handlers

An integration event handler interprets an upstream fact in the consumer's own
model. It may update a consumer-owned projection or durably request a named
owner command. The receiving context owns its translation and recovery policy.

## Creation recipe

1. Add the producer contract family to the consumer service's selection when this is its first use. Generate the plain integration DTO and codec.
2. Create a business-named application handler receiving the typed fact and the narrow output port it needs.
3. For an aggregate change, first create the owner command/handler and private command codec. Inject `CommandBusPort` and map the incoming fact to an owner instruction.
4. Add a stable `command` subscription entry and bind the event/command consumer pair. Acceptance commits evidence, exact command bytes and dispatch before ACK; the command consumer owns later execution.
5. For an explicit browser reaction, inject `RealtimeBusPort`, define the owner read/view mapping and add a separate `realtime` subscription. Loyalty's [RewardIssued handler](../../../services/engagement/src/contexts/loyalty/application/event-handlers/realtime-reward-issued.ts) is the running example.
6. Register the generated subscription and source-event identifiers, validate the complete owner set and start workers only after DI resolution.
7. Test malformed/duplicate input, conflicting source identity, allowed pending state and eventual effect. An enqueued command may still be pending; only its completed transition can confirm changed aggregate state.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Go: record a collected order

Ordering translates Collection's handover fact into its own `RecordCollection`
command. The generated payload supplies the order and customer; composition selects
`event.OrderID` as the command target.

```go source="../../../services/storefront/contexts/ordering/application/eventhandlers/order_collected.go"
// OrderCollectedIntegrationEventHandler durably requests an owner transition before acknowledging the fact.
type OrderCollectedIntegrationEventHandler struct {
	Commands a.CommandBusPort[orderingcontracts.RecordCollectionCommand]
}

func (h OrderCollectedIntegrationEventHandler) Handle(ctx context.Context, m a.Metadata, event model.OrderCollectedIntegrationEvent) (a.Outcome, error) {
	return h.Commands.Enqueue(ctx, m, orderingcontracts.RecordCollectionCommand{CustomerID: event.CustomerID})
}
```

The [durable-command recipe](durable-commands.md) includes its private schema,
codec, manifest, Fx bus binding and paired consumer registration.
[MenuDirectoryProjectionHandler](../../../services/storefront/contexts/ordering/application/projections/menu_published.go)
shows the separate projection-only role.

## Python: request preparation

### Python

Source: [order_placed.py](../../../services/operations/src/operations/contexts/preparation/application/event_handlers/order_placed.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/event_handlers/order_placed.py"
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
```

The handler translates published lines into preparation instructions and enqueues
AcceptOrder. The command consumer later creates the ticket.

## TypeScript: request a loyalty credit

### TypeScript

Source: [order-collected.ts](../../../services/engagement/src/contexts/loyalty/application/event-handlers/order-collected.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/event-handlers/order-collected.ts"
import type { CommandBusPort, Metadata } from '../../../../foundation/application.js';
import type { OrderCollectedIntegrationEvent } from '../../../../contracts/events/generated/collection.js';
import type { CreditCollectionCommand } from '../../../../contracts/requests/generated/loyalty.js';

/** Maps Collection's fact into a durable intent owned by Loyalty. */
export class OrderCollectedIntegrationEventHandler {
  constructor(private commands: CommandBusPort<CreditCollectionCommand>) {}
  handle(m: Metadata, event: OrderCollectedIntegrationEvent) {
    return this.commands.enqueue(m, { orderId: event.orderId, customerId: event.customerId });
  }
}
```

CreditCollection executes later. Its command identity is derived from the order
collection so separate messages describing the same business collection cannot
earn duplicate credit. The original event ID still identifies the received message.

## Declare the receiving boundary

The [subscription manifests](contracts-and-generation.md) name the consumer,
upstream event and private command when applicable. Composition pairs event and
command consumers and checks the complete declared set before workers start.

An event ACK means its owned work is durable. For all three
command examples that work is the receiving evidence plus exact queued command bytes
and dispatch intent. It does not imply the later aggregate command has completed.

## Related concepts

[Integration events](integration-events.md) · [Durable commands](durable-commands.md) · [Projections](projections.md) · [Receipts and idempotency](receipts-and-idempotency.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
