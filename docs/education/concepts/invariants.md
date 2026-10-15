# Invariants and business rejections

An invariant is a business condition that must hold at the transition boundary
where it applies. Aggregate design starts by identifying which state must satisfy
that condition together and which pending states the business permits.

## Guard transitions in the root

### Go

Source: [order.go](../../../services/storefront/contexts/ordering/domain/order.go).

```go source="../../../services/storefront/contexts/ordering/domain/order.go"
func (o *Order) ChangeQuantity(id string, quantity int) error {
	if o.state.Status != "draft" {
		return core.Reject("order_already_placed", "Placed orders cannot be edited")
	}
	q, err := core.NewQuantity(quantity)
	if err != nil {
		return err
	}
	for i, line := range o.lines {
		if line.id == id {
			if o.total()-line.quantity.Value()+q.Value() > 5 {
				return tooManyDrinks()
			}
			if line.quantity.Value() == q.Value() {
				return nil
			}
			line.quantity = q
			o.lines[i] = line
			o.facts = append(o.facts, OrderLineQuantityChangedDomainEvent{State: line.Snapshot()})
			return nil
		}
	}
	return core.Reject("line_not_found", "The order line does not exist")
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

Order checks the total across its children during line edits. An empty draft is
valid during assembly; placement requires at least one drink. A ticket must be
preparing before completion. Reward redemption requires an issued, unexpired
reward and a valid order identity.

## Classify failure deliberately

An expected rule rejection has a stable code and recorded business outcome.
Repeated delivery can recover that outcome. A database outage or corrupt stored
state remains a technical failure eligible for infrastructure retry and diagnosis.
Converting such a failure into a permanent business rejection can lose recoverable work.

The [transaction executor](transactions.md) performs this classification. Domain
code owns the rule and its refusal. Transport code maps the recorded outcome to
the caller's contract without moving policy into HTTP validation.

## Protect against concurrent decisions

Two valid decisions based on the same old state may conflict when combined.
The expected revision and transaction serialisation protect the relationship
between the caller's observation and the write. Domain checks alone cannot
prove that concurrency boundary; run the real database tests.

## Related concepts

[Aggregates](aggregates.md) · [Transactions](transactions.md) · [Receipts and idempotency](receipts-and-idempotency.md) · [Domain layer](domain-layer.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
