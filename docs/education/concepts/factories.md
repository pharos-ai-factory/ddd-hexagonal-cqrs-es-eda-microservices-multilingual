# Factories and creation operations

A factory or named creation operation establishes valid initial aggregate state.
Creation is a business transition. The owning command handler supplies its inputs,
invokes the factory and saves the new root through the normal transaction boundary.

## Creation examples

Go's Order constructor validates identities and currency before returning an empty
draft. Python's AcceptOrder handler constructs a queued ticket from translated
order instructions. TypeScript's IssueReward handler invokes `Reward.issue`.

### Go

Source: [order.go](../../../services/storefront/contexts/ordering/domain/order.go).

```go source="../../../services/storefront/contexts/ordering/domain/order.go"
func New(id, customer, edition, currency string) (*Order, error) {
	for _, value := range []string{id, customer, edition} {
		if err := core.ValidateID(value); err != nil {
			return nil, err
		}
	}
	if _, err := core.NewMoney(0, currency); err != nil {
		return nil, err
	}
	return &Order{state: OrderState{ID: id, CustomerID: customer, EditionID: edition, Currency: currency, Status: "draft", Lines: []LineState{}}}, nil
}
```

### Python

Source: [accept_order.py](../../../services/operations/src/operations/contexts/preparation/application/commands/accept_order.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/commands/accept_order.py"
from operations.contracts.requests.generated.preparation import AcceptOrderCommand
from operations.foundation.application import (
    RealtimeBusPort,
    RealtimeMessage,
    context_target,
    snapshot_message,
)
from operations.foundation.write_repository import CommandResult, CommandContext
from operations.contexts.preparation.application.ports.ticket_write_repository import TicketWriteRepository
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket, TicketState


class AcceptOrderCommandHandler:
    """Create one PreparationTicket; redelivery preserves the existing ticket."""

    def __init__(self, repository: TicketWriteRepository, realtime: RealtimeBusPort[RealtimeMessage]) -> None:
        self.repository = repository
        self.realtime = realtime

    def execute(self, context: CommandContext, command: AcceptOrderCommand) -> CommandResult:
        loaded = self.repository.get(context.target)
        state = loaded["state"] if loaded else None
        if state is not None:
            restored = state.snapshot()
            return CommandResult(restored["status"])
        ticket = PreparationTicket(
            TicketState(context.target, command["orderId"], command["customerId"], command["instructions"])
        )
        self.repository.save(ticket)
        self.realtime.publish(snapshot_message(context_target("preparation"), ticket.snapshot()))
        return CommandResult("queued")
```

### TypeScript

Source: [issue-reward.ts](../../../services/engagement/src/contexts/loyalty/application/commands/issue-reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/commands/issue-reward.ts"
import type { IssueRewardCommand } from '../../../../contracts/requests/generated/loyalty.js';
import type { CommandContext } from '../../../../foundation/write-repository.js';
import type { RewardWriteRepository } from '../ports/reward-write-repository.js';
import { Reward } from '../../domain/reward.js';

/** Carries the immutable grant so reward issuance can recover independently. */

/** Creates one Reward from its immutable earned grant and injected clock. */
export class IssueRewardCommandHandler {
  constructor(
    private repository: RewardWriteRepository,
    private clock: () => Date,
  ) {}
  async execute(context: CommandContext, command: IssueRewardCommand) {
    const loaded = await this.repository.get(context.target);
    if (loaded) {
      const state = loaded.state.snapshot();
      return { status: state.status };
    }
    const reward = Reward.issue(
      context.target,
      {
        id: command.grantId,
        accountId: command.accountId,
        benefit: command.benefit,
        validDays: command.validDays,
      },
      this.clock(),
    );
    await this.repository.save(reward);
    return { status: 'issued' };
  }
}
```

The Python construction belongs inside AcceptOrder's handler even though its
domain type uses a constructor rather than a named static factory. The TypeScript
factory records RewardIssued; its handler stages that root for persistence.

## Creation and restoration have different responsibilities

Creation enforces the initial business rules and may produce new facts.
[Hydration](aggregate-hydration.md) validates existing saved authority and restores
its lifecycle without issuing a new reward or repeating a placement. A constructor
can serve restoration when its semantics are explicitly written for that role.

Use stable target identities so retries address the same root. The central
transaction handles the command identity, expected revision and recorded outcome.
An independent aggregate-changing creation reaction arrives through a durable command.

## Related concepts

[Aggregates](aggregates.md) · [Command handlers](command-handlers.md) · [Aggregate hydration](aggregate-hydration.md) · [Durable commands](durable-commands.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
