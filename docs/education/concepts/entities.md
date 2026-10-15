# Entities

An entity has continuity of identity across changes to its attributes. The
business decides when two representations refer to the same thing. Identity
can survive changes in status, quantity or descriptive fields.

## Identity in the three domain models

OrderLine is a child entity: changing its quantity preserves its line identity.
The Order root controls that change. PreparationTicket and Reward are root
entities with their own repository boundaries. These state shapes retain identity:

### Go

Source: [order.go](../../../services/storefront/contexts/ordering/domain/order.go).

```go source="../../../services/storefront/contexts/ordering/domain/order.go"
type OrderLine struct {
	id        string
	selection Selection
	quantity  core.Quantity
	price     core.Money
}
```

### Python

Source: [preparation_ticket.py](../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py"
class TicketState:
    """Stores the private state restored and changed by a PreparationTicket."""

    id: str
    order_id: str
    customer_id: str
    instructions: str
    status: TicketStatus = "queued"
```

### TypeScript

Source: [reward.ts](../../../services/engagement/src/contexts/loyalty/domain/reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/domain/reward.ts"
export type RewardState = Readonly<{
  id: string;
  grantId: string;
  customerId: string;
  benefit: string;
  status: 'issued' | 'redeemed' | 'expired';
  expiresAt: string;
  redeemedFor?: string;
}>;
```

The Python `TicketState` is the immutable state representation of the entity;
the surrounding `PreparationTicket` provides behaviour. Similarly, TypeScript's
`RewardState` is data used by the Reward aggregate. A record type alone does not
establish the entity's behavioural boundary.

## Identity and equality

OrderLine's `SameIdentity` compares its identifier. Two line snapshots can differ
in quantity and still describe the same line. By contrast, two Money values are
equal when amount and currency match. See [value objects](value-objects.md).

## Decide which entity is a root

Every aggregate root is an entity. An entity inside an aggregate has no independent
write repository here. OrderLine edits go through Order so the total quantity
rule can consider every line. Reward has an independent lifecycle and receives
its own commands. A database primary key alone does not determine that choice.

## Related concepts

[Value objects](value-objects.md) · [Aggregates](aggregates.md) · [Invariants](invariants.md) · [Write repositories](write-repositories.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
