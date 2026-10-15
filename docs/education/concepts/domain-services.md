# Domain services

A domain service expresses business behaviour that has no natural entity or
value-object owner. It uses the domain's language and remains independent of
transport, persistence and dependency-injection frameworks.

## Decide whether the rule already has an owner

| Existing decision                | Owner             | Why                                |
| -------------------------------- | ----------------- | ---------------------------------- |
| Can this order be placed?        | Order             | Depends on its lifecycle and lines |
| Can this ticket become ready?    | PreparationTicket | Depends on its preparation state   |
| Can this reward be redeemed now? | Reward            | Depends on its terms and lifecycle |

These rules are implemented on the roots in Go, Python and TypeScript. Extracting
them into a generic service would separate the rule from the state it protects.
Their actual code is on the [aggregate](aggregates.md) and [invariant](invariants.md) pages.

## A possible new rule

Suppose the business introduced a pricing policy that compares several immutable
offer values and decides an eligible discount. A pure domain service could return
that decision for the owning aggregate to apply. That is a design exercise;
this repository currently has no such pricing service.

The application handler would inject the policy through an appropriate plain
collaborator. Go can use a function or interface, Python a callable or protocol,
and TypeScript a function or interface. Introduce the abstraction when the business
rule justifies it, with tests of its decisions.

## Distinguish neighbouring roles

An application service coordinates loading and saving. A PostgreSQL repository
implements storage. A messaging relay delivers bytes. Their use of the word
“service” does not place them in the domain layer.

A domain service also does not create permission to synchronously mutate several
roots. Immediate cross-root invariants require a boundary decision; permitted
asynchronous work uses an explicit command workflow.

## Inspect the existing rule owners

These actual domain operations show why the current examples keep the rule on
the aggregate. A proposed service should have a distinct business responsibility
that these roots cannot express naturally. No standalone domain-service class is
implemented for these decisions.

### Go: placement rule

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

### Python: completion rule

Source: [preparation_ticket.py](../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py"
def complete(self) -> None:
    if self._state.status != "preparing":
        raise Rejection("preparation_not_started", "Only preparing tickets can become ready")
    self._state = replace(self._state, status="ready")
    self._facts.append(DrinksReadyDomainEvent(self._state.order_id, self._state.customer_id))
```

### TypeScript: redemption rule

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

## A standalone policy in three languages

The following equivalent examples are illustrative extensions, absent from the
running café. Assume validated non-negative minor units and drink counts. The
hypothetical policy awards a ten-percent group discount for three or more drinks;
integer division rounds down. A real requirement would need currency and rounding
rules agreed with domain experts before adoption.

### Go: illustrative policy

```go
type GroupDiscountPolicy struct{}

func (GroupDiscountPolicy) Discount(totalMinor int64, drinks int) int64 {
	if drinks < 3 {
		return 0
	}
	return totalMinor / 10
}
```

### Python: illustrative policy

```python
class GroupDiscountPolicy:
    def discount(self, total_minor: int, drinks: int) -> int:
        return total_minor // 10 if drinks >= 3 else 0
```

### TypeScript: illustrative policy

```ts
class GroupDiscountPolicy {
  discount(totalMinor: number, drinks: number): number {
    return drinks >= 3 ? Math.floor(totalMinor / 10) : 0;
  }
}
```

The policy has no database or transport dependency. An application handler could
supply it to the relevant domain operation; the aggregate would retain responsibility
for applying the decision within its own invariant. The existing rule-owner
excerpts above show the code this extension would need to fit around.

## Related concepts

[Domain layer](domain-layer.md) · [Application layer](application-layer.md) · [Aggregates](aggregates.md) · [Ports](ports.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
