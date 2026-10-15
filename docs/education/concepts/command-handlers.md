# Command handlers

A command handler is an [application-layer](application-layer.md) use case. It
loads its target through a named [write repository](write-repositories.md), asks
the [aggregate](aggregates.md) to decide and stages the changed state.

## Creation recipe

1. Complete the [command contract and registration recipe](commands.md): payload, request envelope, generated identifier and owner parser.
2. Define the command in Protobuf, run generation and import its generated plain input into the named handler in `application/commands/<action>`. Inject the named write repository and any plain clock/identity/projection ports needed by the use case.
3. Load the target through `get`. Handle missing resources explicitly, or create the root when this is a creation command. Follow [hydration](aggregate-hydration.md).
4. Invoke one aggregate behaviour and save its resulting state. The aggregate records its private facts; its owner mapper selects delivered facts.
5. If this use case directly updates the browser, inject `RealtimeBusPort<RealtimeEvent>` and call `publish` with the selected snapshot after saving. For an event-driven update, declare the separate subscriber in the [realtime recipe](realtime-events.md).
6. Register the parser and plain handler factory with the aggregate's command registration. The factory receives the invocation's write repository and realtime bus; the executor owns the transaction and closes both capabilities afterwards.
7. Test absence, invariant rejection, saved state, fact selection, explicit publication and duplicate delivery through the real executor. Expected rejections can publish a matching outcome; infrastructure failures roll back.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## The complete use cases in three languages

These are the repository implementations. The port is a Go field or a constructor
argument in Python and TypeScript. Missing state is an explicit business rejection.
The domain operation owns the rule; the handler coordinates the steps.

### Go

Source: [place_order.go](../../../services/storefront/contexts/ordering/application/commands/place_order.go).

```go source="../../../services/storefront/contexts/ordering/application/commands/place_order.go"
package commands

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	orderingcontracts "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/model/ordering"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// PlaceOrderCommandHandler applies PlaceOrder through one aggregate command transaction.
type PlaceOrderCommandHandler struct {
	Repository ports.OrderWriteRepository
}

func (h PlaceOrderCommandHandler) Execute(ctx context.Context, m a.CommandContext, _ orderingcontracts.PlaceOrderCommand) (a.CommandResult, error) {
	s, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !s.Exists {
		return a.CommandResult{}, orderNotFound()
	}
	order := s.State
	if err = order.Place(); err != nil {
		return a.CommandResult{}, err
	}
	if err = h.Repository.Save(ctx, order); err != nil {
		return a.CommandResult{}, err
	}
	return a.Result("placed"), nil
}
```

### Python

Source: [complete_preparation.py](../../../services/operations/src/operations/contexts/preparation/application/commands/complete_preparation.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/commands/complete_preparation.py"
from operations.contracts.requests.generated.preparation import CompletePreparationCommand
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.preparation.application.ports.ticket_write_repository import TicketWriteRepository
from operations.foundation.domain import Rejection


class CompletePreparationCommandHandler:
    """Completes one ticket and records the outgoing preparation fact atomically."""

    def __init__(self, repository: TicketWriteRepository) -> None:
        self.repository = repository

    def execute(self, context: CommandContext, command: CompletePreparationCommand) -> CommandResult:
        loaded = self.repository.get(context.target)
        state = loaded["state"] if loaded else None
        if state is None:
            raise Rejection("not_found", "The preparation ticket does not exist")
        ticket = state
        ticket.complete()
        self.repository.save(ticket)
        return CommandResult("ready")
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

## Who injects the repository?

The [command registrations](commands.md#6-bind-the-parser-and-handler-in-composition) supply
plain handler factories. The central executor obtains an invocation-scoped
repository from the configured aggregate transaction and passes it to the factory.
Fx, Dependency Injector and Awilix construct the resources and registrations in
the [composition root](dependency-injection.md).

For example, the TypeScript factory is
`(repository) => new RedeemRewardCommandHandler(repository, clock)`. The repository
belongs to that command invocation; the injected clock is a previously resolved
collaborator. Application code imports neither Awilix nor PostgreSQL.

## Load, restore, mutate, save, commit

1. `get` / `Get` asks the port for the target root and its revision.
2. The persistence adaptor [rehydrates](aggregate-hydration.md) the aggregate from
   validated current state. The handler receives the restored object.
3. `Place`, `complete` or `redeem` checks the invariant and changes owned state.
4. `save` / `Save` stages the snapshot. An event-recording wrapper collects selected
   [domain facts](domain-events.md) for private delivery.
5. The [transaction executor](transactions.md) commits state, outcome, receipts
   and outgoing private-event/realtime intent. The consumer acknowledges afterwards.

`Save` is therefore not the commit point. Handlers contain no transaction callbacks,
message acknowledgements or calls to other command handlers. A later aggregate
transition uses a [durable command](durable-commands.md).

Reward redemption changes state without raising a new delivered domain event in
this implementation. Reward issuance demonstrates event creation on the
[aggregate page](aggregates.md). The redemption handler explicitly publishes its changed snapshot through `RealtimeBusPort`.

## Related concepts

[Commands](commands.md) · [Dependency injection](dependency-injection.md) · [Aggregate hydration](aggregate-hydration.md) · [Transactions](transactions.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
