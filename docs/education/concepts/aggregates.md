# Aggregates

An aggregate is the consistency boundary controlled through one root. External
code requests meaningful behaviour from that root, which protects its own state
and any child entities. This repository commits at most one root per command.

## Creation recipe

1. Name the invariant, root and children. Give each runtime business transition an owner command as described in [commands](commands.md).
2. Define the state and a validated construction/restoration path. Keep newly created state distinct from [rehydration](aggregate-hydration.md), which must start with no pending facts.
3. Add behaviour that validates the immediate invariant before changing state. Use typed business rejections for expected refusals.
4. Define a plain immutable domain fact and queue it alongside the successful transition. Record the terms required by downstream consumers at occurrence time.
5. Expose a snapshot for owner persistence and a way to inspect pending facts. Add the named [write repository](write-repositories.md) adaptor that calls these methods.
6. Select any delivered private facts in the owner messaging mapper, then configure the aggregate transaction and its command handlers in composition.
7. Test accepted/rejected transitions and restoration separately. Add an explicit application/event handler for [realtime](realtime-events.md); the aggregate itself receives no transport or bus dependency.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Mutation and domain facts belong together

Go changes an Order to placed and queues its state as an OrderPlaced fact.
Python changes a ticket to ready and queues DrinksReady. TypeScript's creation
operation establishes a Reward and queues RewardIssued. These operations run
inside their named owner command handlers.

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
static issue(id: string, grant: Grant, now: Date): Reward {
  [id, grant.id, grant.accountId].forEach(identifier);
  if (!grant.benefit || !Number.isInteger(grant.validDays) || grant.validDays < 1 || grant.validDays > 30) {
    throw new Rejection('invalid_grant', 'The earned grant has invalid terms');
  }
  const reward = new Reward({
    id,
    grantId: grant.id,
    customerId: grant.accountId,
    benefit: grant.benefit,
    status: 'issued',
    expiresAt: new Date(now.getTime() + grant.validDays * 86400000).toISOString(),
  });
  reward.#events.push({ type: 'RewardIssued', reward: reward.snapshot() });
  return reward;
}
```

The queued fact is plain in-memory data. Persistence later selects which facts
need private delivery. A failure before commit leaves no committed state or
outgoing intent. [Domain events](domain-events.md) explains the selection and the
[outbox](outbox.md) explains durable delivery.

## Root, children and separate lifecycles

OrderLine belongs inside Order because the root protects the combined quantity
and placement lifecycle. MenuOffer belongs inside MenuEdition as frozen terms.
LoyaltyAccount and Reward have separate roots: the account atomically earns a
grant, and issuance may remain pending until another command creates the Reward.

Immediate invariants spanning proposed roots are evidence to reconsider their
boundary. A reference or foreign key by itself does not establish one aggregate.

## Restoration differs from a business operation

Loading restores valid saved state without re-running `issue`, `Place` or `complete`.
Those operations would repeat decisions and facts. The [hydration adaptor](aggregate-hydration.md)
uses the dedicated restoration path, leaving the pending event collection empty.

Snapshots return copies of state. Event accessors similarly protect internal
collections. Directly changing a returned view is not a supported business write.

## Related concepts

[Entities](entities.md) · [Invariants](invariants.md) · [Factories](factories.md) · [Command handlers](command-handlers.md) · [Domain events](domain-events.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
