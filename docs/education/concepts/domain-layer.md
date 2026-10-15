# Domain layer

The domain layer expresses business meaning and protects business rules. Its
objects describe valid state and behaviour independently of HTTP, RabbitMQ,
PostgreSQL and dependency-injection frameworks.

## Decisions in three languages

The following methods belong to aggregate roots. Each uses its current state to
accept or reject a transition. The application handler supplies the loaded root
and calls the operation. The TypeScript method receives time explicitly.

### Go

Source: [order.go](../../../services/storefront/contexts/ordering/domain/order.go).

```go source="../../../services/storefront/contexts/ordering/domain/order.go"
func (o *Order) Place() error {
	if o.state.Status != "draft" {
		return core.Reject("order_already_placed", "The order is already placed")
	}
	if o.total() < 1 {
		return core.Reject("empty_order", "An order must contain at least one drink")
	}
	o.state.Status = "placed"
	o.facts = append(o.facts, OrderPlacedDomainEvent{State: o.Snapshot()})
	return nil
}
```

### Python

Source: [preparation_ticket.py](../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py"
def complete(self) -> None:
    if self._state.status != "preparing":
        raise Rejection("preparation_not_started", "Only preparing tickets can become ready")
    self._state = replace(self._state, status="ready")
    self._facts.append(DrinksReadyDomainEvent(self._state.order_id, self._state.customer_id))
```

### TypeScript

Source: [reward.ts](../../../services/engagement/src/contexts/loyalty/domain/reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/domain/reward.ts"
redeem(orderId: string, now: Date): void {
  identifier(orderId);
  if (this.#state.status !== 'issued') throw new RewardUnavailableDomainError();
  if (now.getTime() >= Date.parse(this.#state.expiresAt))
    throw new Rejection('reward_expired', 'The reward has expired');
  this.#state = { ...this.#state, status: 'redeemed', redeemedFor: orderId };
}
```

## What belongs here

[Entities](entities.md) retain identity, [value objects](value-objects.md) express
validated values, and [aggregates](aggregates.md) protect immediate consistency.
[Factories](factories.md) establish valid initial state. A [domain service](domain-services.md)
can express a rule that has no natural entity or value-object owner.

Plain [domain events](domain-events.md) record selected transitions in the owner's
language. An aggregate queues a fact in memory. It has no broker client and does
not know which integration subscribers exist.

## Boundary checks

Domain code never imports another context. It does not select SQL rows, resolve
containers, acknowledge messages or encode transport envelopes. Those concerns
belong to [adaptors](adaptors.md) and [composition](dependency-injection.md).

Domain tests use deterministic inputs, fixed identifiers and explicit timestamps.
Database tests separately prove that the accepted change and outgoing intent
commit atomically. The two kinds of evidence support different claims.

## Related concepts

[Aggregates](aggregates.md) · [Invariants](invariants.md) · [Domain events](domain-events.md) · [Application layer](application-layer.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
