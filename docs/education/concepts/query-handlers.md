# Query handlers

A query handler coordinates one read use case through a named read repository.
It returns application views and preserves their revision and missing-resource
semantics. It has no aggregate mutation capability.

## Creation recipe

1. Follow the [query creation recipe](queries.md) through schema, envelope, generated identifier and application input.
2. Define the application-owned view and named `<Resource>ReadRepository` capability. Decide missing-resource and pagination behaviour before implementing the handler.
3. Implement the handler as a plain read-repository consumer. Preserve the loaded view's revision and list continuation.
4. Implement the owner request parser and reply mapper. The handler receives its application input and returns its application result.
5. Bind the concrete repository and handler in DI. Register generated ID, parser, handler and reply mapper together; the request registry dispatches through `QueryBusPort`.
6. Test the registered query through its transport boundary as well as read behaviour. Missing roots, corrupt state and incomplete lists must retain distinct outcomes.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Implementations in three languages

### Go

Source: [get_order.go](../../../services/storefront/contexts/ordering/application/queries/get_order.go).

```go source="../../../services/storefront/contexts/ordering/application/queries/get_order.go"
package queries

import (
	"context"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// GetOrderQuery identifies the order requested by the caller.
type GetOrderQuery struct{ ID string }

// GetOrderQueryHandler reads one order through its application capability.
type GetOrderQueryHandler struct{ ReadRepository ports.OrderReadRepository }

func (h GetOrderQueryHandler) Execute(ctx context.Context, q GetOrderQuery) (a.Loaded[view.OrderView], error) {
	return h.ReadRepository.Get(ctx, q.ID)
}
```

### Python

Source: [get_ticket.py](../../../services/operations/src/operations/contexts/preparation/application/queries/get_ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/queries/get_ticket.py"
from dataclasses import dataclass
from operations.foundation.application import Loaded
from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository
from operations.contexts.preparation.application.read_models.ticket import TicketView


@dataclass(frozen=True)
class GetTicketQuery:
    """Select one ticket."""

    identity: str


class GetTicketQueryHandler:
    """Execute the GetTicket read use case through its owning application port."""

    def __init__(self, read_repository: TicketReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: GetTicketQuery) -> Loaded[TicketView] | None:
        return self.read_repository.get(query.identity)
```

### TypeScript

Source: [get-reward.ts](../../../services/engagement/src/contexts/loyalty/application/queries/get-reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/queries/get-reward.ts"
import type { RewardReadRepository } from '../ports/reward-read-repository.js';

/** Select one reward. */
export type GetRewardQuery = Readonly<{ id: string }>;

/** Execute the GetReward use case through its context-owned read port. */
export class GetRewardQueryHandler {
  constructor(private readRepository: RewardReadRepository) {}
  async execute(query: GetRewardQuery) {
    return this.readRepository.get(query.id);
  }
}
```

These handlers are intentionally small. The [read adaptor](read-repositories.md)
validates stored authority and creates the view; the transport reply mapper turns
that result into the published reply. Keeping those roles explicit allows each
boundary to be tested independently.

## Inject the read repository

### Go

Source: [ordering.go](../../../services/storefront/apps/storefront/di/ordering.go).

```go source="../../../services/storefront/apps/storefront/di/ordering.go"
func(read ports.OrderReadRepository) q.GetOrderQueryHandler {
	return q.GetOrderQueryHandler{ReadRepository: read}
},
```

### Python

Source: [preparation.py](../../../services/operations/src/operations/apps/composition/preparation.py).

```python source="../../../services/operations/src/operations/apps/composition/preparation.py"
ticket_read_repository = providers.Singleton(PostgresTicketReadRepository, database)
get_ticket = providers.Singleton(GetTicketQueryHandler, ticket_read_repository)
list_tickets = providers.Singleton(ListTicketsQueryHandler, ticket_read_repository)
```

### TypeScript

Source: [loyalty.ts](../../../services/engagement/src/apps/composition/loyalty.ts).

```ts source="../../../services/engagement/src/apps/composition/loyalty.ts"
rewardReadRepository: asFunction(
  (c: LoyaltyDependencies) => new PostgresRewardReadRepository(c.database),
).singleton(),
getReward: asFunction(
  (c: LoyaltyDependencies) => new GetRewardQueryHandler(c.rewardReadRepository),
).singleton(),
listRewards: asFunction(
  (c: LoyaltyDependencies) => new ListRewardsQueryHandler(c.rewardReadRepository),
).singleton(),
```

Composition resolves the handler before registering its query. The repository
retains a pool-backed read capability rather than a command's transaction-scoped
write repository. These lifetimes are deliberately different.

An ordinary query may return a projection that lags behind another context.
State that freshness property explicitly. The handler should not call another
context's private database to hide the delay.

## Related concepts

[Queries](queries.md) · [Application layer](application-layer.md) · [Read repositories](read-repositories.md) · [Dependency injection](dependency-injection.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
