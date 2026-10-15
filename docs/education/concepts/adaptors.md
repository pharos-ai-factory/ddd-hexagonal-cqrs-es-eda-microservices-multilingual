# Adaptors

An adaptor connects a [port](ports.md) to a concrete technology and translates
data at that boundary. Driving adaptors invoke application operations. Driven
adaptors implement capabilities the application needs.

## Incoming command mapping

The same boundary appears in the three services: generated transport values are
validated and translated into plain application commands.

### Go

Source: [ordering_requests.go](../../../services/storefront/contexts/ordering/adaptors/messaging/ordering_requests.go).

```go source="../../../services/storefront/contexts/ordering/adaptors/messaging/ordering_requests.go"
func PlaceOrder(p *pb.PlaceOrder) commands.PlaceOrderCommand {
	return commands.PlaceOrderCommand{}
}
```

### Python

Source: [ticket_requests.py](../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py).

```python source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py"
def complete(value: Message) -> CompletePreparationCommand:
    if not isinstance(value, CompletePreparation):
        raise ValueError("Unexpected complete preparation payload")
    return CompletePreparationCommand()
```

### TypeScript

Source: [reward-requests.ts](../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-requests.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-requests.ts"
export function redeem(value: unknown): { orderId: string } {
  const fields = record(value);
  const wire: wire.RedeemReward = { orderId: text(fields.orderId) };
  return { orderId: identifier(wire.orderId!) };
}
```

These functions are passed to the [command registration](commands.md). Wire
validation handles message shape; aggregate behaviour still enforces business
invariants after the handler has loaded authoritative state.

## Outgoing and persistence boundaries

Persistence adaptors restore aggregates, map read models and stage snapshots.
Messaging adaptors encode private/public contracts and implement broker delivery.
Realtime adaptors encode browser projections and publish through the gateway.
Operational HTTP adaptors expose health and authenticated diagnostics.

The public business HTTP boundary is the separate Go API. Owner services receive
business commands and queries through messaging. A browser or another context
does not call owner business HTTP routes.

Context-specific mappings belong with their owning context. Shared service
adaptors provide technical transaction and delivery mechanics. See the
[adaptor layout](../../adaptor-layout.md) for the current file map.

## Related concepts

[Ports](ports.md) · [Commands](commands.md) · [Aggregate hydration](aggregate-hydration.md) · [Realtime events](realtime-events.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
