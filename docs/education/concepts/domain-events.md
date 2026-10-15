# Domain events

A domain event records something that happened in its owner's model. It is a
plain fact raised alongside an accepted aggregate transition. Delivery identity,
transport encoding and subscriber configuration belong outside the aggregate.

## Creation recipe

1. Define the plain owner `*DomainEvent` and queue it in the aggregate method after its invariant succeeds. See the [aggregate recipe](aggregates.md).
2. Add its private payload and envelope alternative to the owner's `adaptors/messaging/private_messages.proto`; preserve existing field numbers and stored-byte compatibility.
3. Add the stable private event identity, owner, aggregate kind and visibility to `private_messages.json`. Domain events remain private even when delivered through RabbitMQ.
4. Implement the owner private codec and semantic validation. Add an encoded fixture under that adaptor's `fixtures/` and test both decoding and metadata consistency.
5. Extend the aggregate-to-private-publication mapper shown below. Configure that mapper on the aggregate transaction; its `DomainEventBusPort` stages selected facts with the root.
6. Add each desired subscriber to `subscriptions.json`: `publication` for integration mapping, `command` for durable owner reactions, or `realtime` for browser publication.
7. Run `pnpm generate:contracts`, bind generated names and each plain handler in composition, validate the complete subscription set and test delivery/retry behaviour. The [domain-handler recipe](domain-event-handlers.md) continues from the delivered fact.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Define a plain fact

### Go

Source: [events.go](../../../services/storefront/contexts/ordering/domain/events.go).

```go source="../../../services/storefront/contexts/ordering/domain/events.go"
// OrderPlacedDomainEvent records the completed OrderPlaced transition.
type OrderPlacedDomainEvent struct{ State OrderState }

func (OrderPlacedDomainEvent) DomainEvent() {}
```

### Python

Source: [preparation_ticket.py](../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py"
class DrinksReadyDomainEvent:
    """Private fact raised when this context completes a preparation ticket."""

    name: ClassVar[str] = "DrinksReady"
    order_id: str
    customer_id: str
```

### TypeScript

Source: [reward.ts](../../../services/engagement/src/contexts/loyalty/domain/reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/domain/reward.ts"
export type RewardIssuedDomainEvent = Readonly<{ type: 'RewardIssued'; reward: RewardState }>;
```

The Python declaration has `@dataclass(frozen=True)` in its source. Go uses a
marker method; TypeScript uses a discriminated read-only object. Each represents
owner data, independently of the generated public integration DTO.

The [aggregate examples](aggregates.md#mutation-and-domain-facts-belong-together)
show the exact operation that changes state and appends each fact. Restoring a
saved root starts with no newly pending facts.

## Select facts for private delivery

Saving uses these owner mappings. They select facts that have asynchronous
subscribers and create private publication intent. The central transaction writes
their exact encoded bytes with the root.

### Go

Source: [order_domain_events.go](../../../services/storefront/contexts/ordering/adaptors/messaging/publications/order_domain_events.go).

```go source="../../../services/storefront/contexts/ordering/adaptors/messaging/publications/order_domain_events.go"
// Package publications stages owner-private domain events.
package publications

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	names "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/names"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// Order selects committed facts for independent private subscriptions.
func Order(aggregate *d.Order) []a.Publication {
	var result []a.Publication
	for _, fact := range aggregate.Events() {
		if event, ok := fact.(d.OrderPlacedDomainEvent); ok {
			result = append(result, a.Publication{Name: names.OrderingDomainEventOrderPlaced.String(), Visibility: a.Private, Payload: event})
		}
	}
	return result
}
```

### Python

Source: [ticket_publications.py](../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_publications.py).

```python source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_publications.py"
"""Stage Preparation's private facts in the aggregate transaction."""

from operations.adaptors.generated.message_names import PreparationDomainEvent

from operations.contexts.preparation.domain.preparation_ticket import (
    PreparationTicket,
    DrinksReadyDomainEvent,
)
from operations.foundation.application import Publication


def ticket_publications(aggregate: PreparationTicket) -> tuple[Publication, ...]:
    """Select the private fact consumed by the durable integration publication handler."""
    return tuple(
        Publication(
            PreparationDomainEvent.DRINKS_READY,
            {"orderId": fact.order_id, "customerId": fact.customer_id},
        )
        for fact in aggregate.events()
        if isinstance(fact, DrinksReadyDomainEvent)
    )
```

### TypeScript

Source: [reward-publications.ts](../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-publications.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-publications.ts"
import { LoyaltyDomainEvent } from '../../../../adaptors/generated/message-names.js';
import type { Reward } from '../../domain/reward.js';
import type { Publication } from '../../../../foundation/application.js';

/** Map the saved aggregate's private facts to this owner's delivered messages. */
export function rewardPublications(aggregate: Reward): Publication[] {
  return aggregate.events().map((fact) => ({
    name: LoyaltyDomainEvent.REWARD_ISSUED,
    payload: {
      rewardId: fact.reward.id,
      customerId: fact.reward.customerId,
      benefit: fact.reward.benefit,
      expiresAt: fact.reward.expiresAt,
    },
  }));
}
```

OrderLineAdded and PreparationStarted are domain facts too, but these mappings
do not send them to the broker. A domain fact does not automatically require a
public contract or any asynchronous delivery.

## Preserve privacy and compatibility

Private codecs and fixtures live in the owner's messaging adaptors. They still
need compatibility discipline because committed bytes may be delivered after a
software update. Generated private routing names identify the envelope; the
domain continues to use its own plain fact types.

An aggregate save may stage only private domain events. A separately delivered
[domain event handler](domain-event-handlers.md) decides whether to publish an
integration fact or enqueue another owner command.

## Related concepts

[Aggregates](aggregates.md) · [Write repositories](write-repositories.md) · [Domain event handlers](domain-event-handlers.md) · [Outbox](outbox.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
