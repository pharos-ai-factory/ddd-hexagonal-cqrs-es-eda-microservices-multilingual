# Domain event handlers

A domain event handler reacts to a committed private fact. Publication handlers
translate that fact into a public integration DTO. Aggregate-changing reactions
enqueue a named owner command through a durable command port.

## Creation recipe

1. Complete the [private domain-event recipe](domain-events.md), including schema, catalogue, codec and compatibility fixture.
2. Give the reaction its own business-named application file. Choose its output explicitly: `IntegrationEventBusPort`, `CommandBusPort` or `RealtimeBusPort`.
3. For integration publication, define the annotated owner Protobuf payload and generate its plain DTO/factory. Map the private fact into that DTO in the handler examples below.
4. Add a stable `publication` subscription declaration linking the private source name to its integration output. A realtime reaction instead has its own `realtime` declaration and handler.
5. Bind the concrete bus/outbox and plain handler in owner DI. Register the private input parser/target and generated subscription identifier with the shared subscription helper.
6. Commit the receiving receipt with exact outgoing bytes and dispatch before ACK. Preserve source revision/time and private-event causation for integration publication.
7. Test concurrent duplicate delivery, conflicting source bytes, failure before receipt commit and recovery after lost ACK. Follow [realtime events](realtime-events.md) for browser-specific snapshot revision rules.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Map domain facts to integration payloads

These handlers are the business mapping between the private and public models.
The output types are generated from published schemas. The mapping itself is
handwritten because the owner decides which facts form the contract.

### Go

Source: [publish_order_placed.go](../../../services/storefront/contexts/ordering/application/eventhandlers/publish_order_placed.go).

```go source="../../../services/storefront/contexts/ordering/application/eventhandlers/publish_order_placed.go"
// Package eventhandlers reacts to committed private facts.
package eventhandlers

import (
	"context"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderPlacedDomainEventHandler publishes the integration contract through an atomic receipt/outbox port.
type OrderPlacedDomainEventHandler struct {
	Publications a.IntegrationEventBusPort[model.OrderPlacedIntegrationEvent]
}

func (h OrderPlacedDomainEventHandler) Handle(ctx context.Context, metadata a.Metadata, event d.OrderPlacedDomainEvent) (a.Outcome, error) {
	s := event.State
	payload := model.OrderPlacedIntegrationEvent{OrderID: s.ID, CustomerID: s.CustomerID, EditionID: s.EditionID, Currency: s.Currency, Lines: []model.Line{}}
	for _, line := range s.Lines {
		payload.Lines = append(payload.Lines, model.Line{ID: line.ID, OfferCode: line.Selection.OfferCode, Name: line.Selection.Name, Quantity: line.Quantity, Minor: line.Selection.Minor})
	}
	return h.Publications.Publish(ctx, metadata, payload)
}
```

### Python

Source: [publish_drinks_ready.py](../../../services/operations/src/operations/contexts/preparation/application/event_handlers/publish_drinks_ready.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/event_handlers/publish_drinks_ready.py"
"""Publish preparation's integration contract after its domain event is committed."""

from operations.contexts.preparation.domain.preparation_ticket import DrinksReadyDomainEvent
from operations.contracts.events.generated.preparation import DrinksReadyIntegrationEvent
from operations.foundation.application import IntegrationEventBusPort, Metadata, Outcome


class DrinksReadyDomainEventHandler:
    """Translate a committed private fact through the atomic receipt/publication port."""

    def __init__(self, publications: IntegrationEventBusPort[DrinksReadyIntegrationEvent]) -> None:
        self.publications = publications

    def handle(self, metadata: Metadata, event: DrinksReadyDomainEvent) -> Outcome:
        return self.publications.publish(
            metadata,
            DrinksReadyIntegrationEvent(
                orderId=event.order_id,
                customerId=event.customer_id,
            ),
        )
```

### TypeScript

Source: [publish-reward-issued.ts](../../../services/engagement/src/contexts/loyalty/application/event-handlers/publish-reward-issued.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/event-handlers/publish-reward-issued.ts"
import type { RewardIssuedDomainEventPayload } from '../events.js';
import type { RewardIssuedIntegrationEvent } from '../../../../contracts/events/generated/loyalty.js';
import type { IntegrationEventBusPort, Metadata } from '../../../../foundation/application.js';

/** Publish the committed reward terms through the atomic receipt/outbox port. */
export class RewardIssuedDomainEventHandler {
  constructor(private publications: IntegrationEventBusPort<RewardIssuedIntegrationEvent>) {}
  handle(metadata: Metadata, event: RewardIssuedDomainEventPayload) {
    return this.publications.publish(metadata, {
      rewardId: event.rewardId,
      customerId: event.customerId,
      benefit: event.benefit,
      expiresAt: event.expiresAt,
    });
  }
}
```

`Publish` / `publish` commits the incoming subscription receipt, exact integration
bytes and dispatch intent atomically. It preserves the source aggregate revision
and occurrence time; causation identifies the private event. The consumer ACKs
after that commit. The integration relay publishes independently afterwards.

## Register the private subscription

The owner subscription manifest declares which private event produces which
public event. Composition supplies the handler and a publication outbox using
the generated typed publication factory. The [integration event page](integration-events.md)
shows that wiring in all three languages.

## A private reaction that changes another aggregate

Loyalty's RewardEarned reaction queues IssueReward. The event handler does not
load or mutate Reward itself:

### TypeScript

Source: [reward-earned.ts](../../../services/engagement/src/contexts/loyalty/application/event-handlers/reward-earned.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/event-handlers/reward-earned.ts"
import type { CommandBusPort, Metadata } from '../../../../foundation/application.js';
import type { RewardEarnedDomainEventPayload } from '../events.js';
import type { IssueRewardCommand } from '../../../../contracts/requests/generated/loyalty.js';

/** Hands an immutable grant to the independent Reward aggregate command queue. */
export class RewardEarnedDomainEventHandler {
  constructor(private commands: CommandBusPort<IssueRewardCommand>) {}
  handle(m: Metadata, event: RewardEarnedDomainEventPayload) {
    return this.commands.enqueue(m, {
      grantId: event.grantId,
      accountId: event.accountId,
      benefit: event.benefit,
      validDays: event.validDays,
    });
  }
}
```

The [durable command](durable-commands.md) consumer later invokes IssueReward's
handler in a separate transaction. An earned grant awaiting issuance is a valid
pending state; retry ownership transfers with the committed hand-off.

## Related concepts

[Domain events](domain-events.md) · [Integration events](integration-events.md) · [Durable commands](durable-commands.md) · [Receipts and idempotency](receipts-and-idempotency.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
