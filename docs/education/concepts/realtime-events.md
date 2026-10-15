# Realtime events: explicitly publish a browser update

Realtime messages are selected by their owning context. A command handler, domain
event handler or integration event handler receives `RealtimeBusPort` and calls
`publish`. The owner adaptor supplies durable storage and authorised delivery.
Saving an aggregate and publishing an integration event each require a separate,
explicit realtime decision. See [decision 0026](../../decisions/0026-explicit-bus-ports-and-realtime-publication.md).

## 1. Choose the producer and the committed fact it describes

Choose one of these existing paths for the new use case:

| Producer                  | Use it when                                                | Transaction                                             |
| ------------------------- | ---------------------------------------------------------- | ------------------------------------------------------- |
| Command handler           | The use case selects a local update or rejection           | Root/outcome, receipt and browser bytes commit together |
| Domain event handler      | An owner-private fact triggers a browser update            | Subscriber receipt and browser bytes commit together    |
| Integration event handler | A declared published fact triggers a consumer-owned update | Subscriber receipt and browser bytes commit together    |

An integration reaction which changes an aggregate enqueues its owner command.
The resulting transition supplies the completed state. A browser refresh handler
may read an owner view; it must carry the revision returned with that view.

## 2. Define the browser contract and audience

Edit `contracts/<context>/realtime/v1/*_snapshots.proto` for the owner payload and
[the separate envelope](../../../contracts/shared/realtime/v1/realtime.proto)
when adding a new message alternative. Preserve field numbers. Run
`pnpm generate:contracts`; inspect all three runtime bindings and the web binding.

The current development interface uses server-authorised `cafe:<owner>` channels.
The concrete bus binds that owner, resource kind and publisher credential. A new
recipient model requires an explicit server-side authorisation policy and gateway
binding. Browser clients never choose their own channels.

## 3. Publish an application update through the injected bus

Add the bus to the command handler constructor/struct. Its registration factory
receives an invocation-scoped bus alongside the repository. Save the root, then
publish its selected snapshot. A rejection can explicitly publish its stable
business code before being returned to the command executor.

These complete handlers demonstrate publication and rejection in every language:

### Go

Source: [change_quantity.go](../../../services/storefront/contexts/ordering/application/commands/change_quantity.go).

```go source="../../../services/storefront/contexts/ordering/application/commands/change_quantity.go"
package commands

import (
	"context"
	"errors"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

// ChangeQuantityCommandHandler applies ChangeQuantity through one aggregate command transaction.
type ChangeQuantityCommandHandler struct {
	Realtime   a.RealtimeBusPort[a.RealtimeEvent]
	Repository ports.OrderWriteRepository
}

func (h ChangeQuantityCommandHandler) Execute(ctx context.Context, m a.CommandContext, c orderingcontracts.ChangeQuantityCommand) (result a.CommandResult, err error) {
	defer func() {
		var rejected d.ExpectedError
		if errors.As(err, &rejected) {
			detail := rejected.Rejection()
			if publishErr := h.Realtime.Publish(ctx, a.RealtimeEvent{Rejection: &detail}); publishErr != nil {
				err = publishErr
			}
		}
	}()
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, orderNotFound()
	}
	order := s.State
	if err = order.ChangeQuantity(c.LineID, c.Quantity); err != nil {
		return a.CommandResult{}, err
	}
	if len(order.Events()) > 0 {
		if err = h.Repository.Save(ctx, order); err != nil {
			return a.CommandResult{}, err
		}
		if err = h.Realtime.Publish(ctx, a.RealtimeEvent{Snapshot: order.Snapshot()}); err != nil {
			return a.CommandResult{}, err
		}
	}
	return a.Result("draft"), nil
}
```

### Python

Source: [collect_order.py](../../../services/operations/src/operations/contexts/collection/application/commands/collect_order.py).

```python source="../../../services/operations/src/operations/contexts/collection/application/commands/collect_order.py"
from operations.contracts.requests.generated.collection import CollectOrderCommand
from operations.foundation.domain import Rejection
from operations.foundation.application import RealtimeBusPort, RealtimeEvent
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.collection.application.ports.pickup_write_repository import PickupWriteRepository
from operations.foundation.application import ApplicationError


class PickupNotFoundApplicationError(ApplicationError):
    """Reports the expected rejection when the targeted pickup has no stored state."""

    def __init__(self) -> None:
        super().__init__("not_found", "The pickup does not exist")


class CollectOrderCommandHandler:
    """Checks the collection code through the aggregate and commits its collection fact."""

    def __init__(self, repository: PickupWriteRepository, realtime: RealtimeBusPort[RealtimeEvent]) -> None:
        self.repository = repository
        self.realtime = realtime

    def execute(self, context: CommandContext, command: CollectOrderCommand) -> CommandResult:
        try:
            loaded = self.repository.get(context.target)
            state = loaded["state"] if loaded else None
            if state is None:
                raise PickupNotFoundApplicationError()
            pickup = state
            pickup.collect(command["code"])
            self.repository.save(pickup)
            self.realtime.publish(RealtimeEvent(snapshot=pickup.snapshot()))
            return CommandResult("collected")
        except Rejection as rejection:
            self.realtime.publish(RealtimeEvent(rejection=rejection.outcome()))
            raise
```

### TypeScript

Source: [redeem-reward.ts](../../../services/engagement/src/contexts/loyalty/application/commands/redeem-reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/commands/redeem-reward.ts"
import type { RedeemRewardCommand } from '../../../../contracts/requests/generated/loyalty.js';
import { Rejection } from '../../../../foundation/domain.js';
import type { RealtimeBusPort, RealtimeEvent } from '../../../../foundation/application.js';
import type { CommandContext } from '../../../../foundation/write-repository.js';
import type { RewardWriteRepository } from '../ports/reward-write-repository.js';
import { ApplicationError } from '../../../../foundation/application.js';

/** Applies the reward redemption rule using the injected clock and one aggregate transaction. */
export class RedeemRewardCommandHandler {
  constructor(
    private repository: RewardWriteRepository,
    private realtime: RealtimeBusPort<RealtimeEvent>,
    private clock: () => Date,
  ) {}
  async execute(context: CommandContext, command: RedeemRewardCommand) {
    try {
      const loaded = await this.repository.get(context.target);
      if (!loaded) throw new RewardNotFoundApplicationError();
      const reward = loaded.state;
      reward.redeem(command.orderId, this.clock());
      await this.repository.save(reward);
      await this.realtime.publish({ snapshot: reward.snapshot() });
      return { status: 'redeemed' };
    } catch (error) {
      if (error instanceof Rejection) await this.realtime.publish({ rejection: error.outcome() });
      throw error;
    }
  }
}

/** Reports the expected rejection when the targeted reward has no stored state. */
export class RewardNotFoundApplicationError extends ApplicationError {
  constructor() {
    super('not_found', 'The reward does not exist');
  }
}
```

The executor copies messages into a scoped buffer. It closes the buffer after
invocation. Infrastructure validates snapshot bytes against the saved state and
rejection intent against the recorded outcome. A failed transaction emits nothing.
An expected rejection discards any staged snapshot update.

## 4. Or create a named event handler for the publication

Give the browser reaction its own file and inject its read repository and bus.
The Go/Python examples consume private domain events. TypeScript deliberately
consumes a published integration event. All three use the revision of the view
they actually read; delayed delivery can therefore refresh a later committed state.

### Go

Source: [realtime_order_placed.go](../../../services/storefront/contexts/ordering/application/eventhandlers/realtime_order_placed.go).

```go source="../../../services/storefront/contexts/ordering/application/eventhandlers/realtime_order_placed.go"
package eventhandlers

import (
	"context"
	"fmt"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderPlacedRealtimeDomainEventHandler publishes the current owner view with its actual revision.
type OrderPlacedRealtimeDomainEventHandler struct {
	Orders   ports.OrderReadRepository
	Realtime a.RealtimeBusPort[a.RealtimeEvent]
}

func (h OrderPlacedRealtimeDomainEventHandler) Handle(ctx context.Context, m a.Metadata, event d.OrderPlacedDomainEvent) (a.Outcome, error) {
	if event.State.ID != m.AggregateID {
		return a.Outcome{}, fmt.Errorf("order event target mismatch")
	}
	loaded, err := h.Orders.Get(ctx, m.AggregateID)
	if err != nil {
		return a.Outcome{}, err
	}
	if !loaded.Exists {
		return a.Outcome{}, fmt.Errorf("committed order is missing")
	}
	err = h.Realtime.Publish(ctx, a.RealtimeEvent{Source: &m, Revision: loaded.Version, Snapshot: loaded.State})
	return a.Outcome{AggregateID: m.AggregateID, Version: loaded.Version, Status: "realtime_queued"}, err
}
```

### Python

Source: [realtime_drinks_ready.py](../../../services/operations/src/operations/contexts/preparation/application/event_handlers/realtime_drinks_ready.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/event_handlers/realtime_drinks_ready.py"
"""Explicit browser publication after the preparation fact has been delivered."""

from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository
from operations.contexts.preparation.domain.preparation_ticket import DrinksReadyDomainEvent
from operations.foundation.application import Metadata, Outcome, RealtimeBusPort, RealtimeEvent


class DrinksReadyRealtimeDomainEventHandler:
    """Publish the current ticket view with its own revision, including after delayed delivery."""

    def __init__(self, tickets: TicketReadRepository, realtime: RealtimeBusPort[RealtimeEvent]) -> None:
        self.tickets, self.realtime = tickets, realtime

    def handle(self, metadata: Metadata, event: DrinksReadyDomainEvent) -> Outcome:
        loaded = self.tickets.get(metadata.target)
        if loaded is None or loaded["state"]["orderId"] != event.order_id:
            raise ValueError("Committed preparation ticket is missing or inconsistent")
        self.realtime.publish(
            RealtimeEvent(source=metadata, revision=loaded["version"], snapshot=loaded["state"])
        )
        return Outcome(aggregateId=metadata.target, version=loaded["version"], status="realtime_queued")
```

### TypeScript

Source: [realtime-reward-issued.ts](../../../services/engagement/src/contexts/loyalty/application/event-handlers/realtime-reward-issued.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/event-handlers/realtime-reward-issued.ts"
import type { RewardReadRepository } from '../ports/reward-read-repository.js';
import type { RewardIssuedIntegrationEvent } from '../../../../contracts/events/generated/loyalty.js';
import type {
  Metadata,
  Outcome,
  RealtimeBusPort,
  RealtimeEvent,
} from '../../../../foundation/application.js';

/** Explicitly translate a published reward fact into an owner-authorised browser update. */
export class RewardIssuedRealtimeIntegrationEventHandler {
  constructor(
    private rewards: RewardReadRepository,
    private realtime: RealtimeBusPort<RealtimeEvent>,
  ) {}
  async handle(metadata: Metadata, event: RewardIssuedIntegrationEvent): Promise<Outcome> {
    if (event.rewardId !== metadata.target) throw new Error('Reward event target mismatch');
    const loaded = await this.rewards.get(metadata.target);
    if (!loaded) throw new Error('Committed reward is missing');
    await this.realtime.publish({ source: metadata, revision: loaded.version, snapshot: loaded.state });
    return { aggregateId: metadata.target, version: loaded.version, status: 'realtime_queued' };
  }
}
```

## 5. Declare and bind the event subscription

Add an entry to the owner's `adaptors/messaging/subscriptions.json`, with a stable
consumer identity, its source event and a `realtime` resource. Existing entries:

### Go

```json source="../../../services/storefront/contexts/ordering/adaptors/messaging/subscriptions.json"
{
  "consumer": "ordering.realtime-order-placed",
  "event": "ordering.order-placed.domain",
  "realtime": "order"
}
```

### Python

```json source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/subscriptions.json"
{
  "consumer": "preparation.realtime-drinks-ready",
  "event": "preparation.drinks-ready.domain",
  "realtime": "ticket"
}
```

### TypeScript

```json source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/subscriptions.json"
{
  "consumer": "loyalty.realtime-reward-issued",
  "event": "loyalty.reward-issued",
  "realtime": "reward"
}
```

Run `pnpm generate:contracts` again. It generates the subscription identifier,
queue bindings and broker permissions. Then wire the concrete bus and handler:

- Go: [Ordering composition](../../../services/storefront/apps/storefront/di/ordering.go) constructs `NewRealtimeEventBus` and binds the generated subscription with `SubscribePrivate`.
- Python: [Preparation composition](../../../services/operations/src/operations/apps/composition/preparation.py) resolves `PostgresRealtimeEventBus` and calls `realtime_subscription`.
- TypeScript: [Loyalty composition](../../../services/engagement/src/apps/composition/loyalty.ts) resolves `PostgresRealtimeEventBus` and calls `realtimeSubscription`.

Include the new worker in the complete owner subscription set. Startup rejects
missing/mismatched declarations. Apply development migrations and generated broker
permissions with `pnpm dev:up` before testing the new runtime.

## 6. Preserve transaction and delivery guarantees

The event bus commits its receipt, exact encoded browser bytes and dispatch intent
before returning to the consumer. ACK follows that commit. Redelivery with identical
source bytes reuses its receipt; conflicting bytes fail. Independent subscriptions
have independent receipts. Each realtime subscription commits one browser
publication per source event; declare separate subscriptions for independently
retryable outputs. The existing relays publish stored bytes through the
credential-scoped gateway to Centrifugo and retry failed delivery.

The concrete implementations are [Go](../../../services/storefront/adaptors/persistence/postgres/realtime_event_bus.go),
[Python](../../../services/operations/src/operations/adaptors/persistence/realtime_event_bus.py)
and [TypeScript](../../../services/engagement/src/adaptors/persistence/realtime-event-bus.ts).
Command publications use the existing aggregate transaction and its invocation
buffer instead. Neither path performs network publication inside a database transaction.

## 7. Decode, apply and recover in the frontend

Update the [browser decoder](../../../services/web/src/adaptors/realtime/projection-codec.ts)
and [realtime hook](../../../services/web/src/shared/realtime/useCafe.ts) when adding a
payload. Snapshot identity and owner must match the envelope/channel. Apply only
newer root revisions. `OperationRejected` carries command/correlation identity and
safe code/message; it is displayed separately and does not replace root state.

Initial attachment and history gaps use [owner queries](queries.md). Each window
maintains its own recovery cursor. Keep uncertain command recovery and session
revocation intact. The request reply and identical command retry remain available
for outcome recovery; a timeout never proves business rejection.

## 8. Verify the declaration and failure paths

Test no publication without a call, copied payloads, expired invocation scope,
encoding rollback, exact bytes, duplicate delivery, conflicting source identity,
late events, business rejections, and browser revision handling. Executable evidence:
[Go](../../../services/storefront/adaptors/persistence/postgres/explicit_realtime_integration_test.go),
[Python](../../../services/operations/tests/realtime_bus_integration.py),
[TypeScript](../../../services/engagement/src/adaptors/persistence/realtime-bus.integration.ts)
and [browser decoding](../../../services/web/src/adaptors/realtime/projection-codec.test.ts).

Run the focused owner tests, `pnpm check --changed`, `pnpm verify` and
`pnpm test:integration`. Continue with the [end-to-end recipe](../end-to-end-recipe.md),
[domain handlers](domain-event-handlers.md), [integration handlers](integration-event-handlers.md)
and [outbox](outbox.md).
