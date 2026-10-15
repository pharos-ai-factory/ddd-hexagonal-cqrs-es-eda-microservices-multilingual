# Query handlers

A query handler coordinates one read use case through a named read repository.
It returns generated response DTOs and preserves their revision and missing-resource
semantics. It has no aggregate mutation capability.

## Creation recipe

1. Follow the [query creation recipe](queries.md) through schema, envelope, generated identifier and generated plain input.
2. Generate the resource DTO and use it at the named `<Resource>ReadRepository` port. Decide missing-resource and pagination behaviour before implementing the handler.
3. Implement the handler as a plain read-repository consumer. Preserve the loaded view's revision and list continuation.
4. Use the generated request parser, retain owner semantic validation and use the generated response encoder. The handler receives its generated plain input and returns its generated response.
5. Bind the concrete repository and handler in DI. Call the generated typed registration with the parser and handler; the request registry dispatches through `QueryBusPort`.
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
	model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"
	shared "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/shared"
)

// GetOrderQueryHandler preserves revision, absence and pagination policy.
type GetOrderQueryHandler struct{ ReadRepository ports.OrderReadRepository }

func (h GetOrderQueryHandler) Execute(ctx context.Context, q model.GetOrderQuery) (model.GetOrderResponse, error) {
	value, err := h.ReadRepository.Get(ctx, q.ID)
	if err != nil {
		return model.GetOrderResponse{}, err
	}
	if !value.Exists {
		return model.GetOrderResponse{Error: &shared.RequestError{Code: "not_found"}}, nil
	}
	return model.GetOrderResponse{Order: &model.LoadedOrder{Exists: value.Exists, Version: value.Version, State: value.State}}, nil
}
```

### Python

Source: [get_ticket.py](../../../services/operations/src/operations/contexts/preparation/application/queries/get_ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/queries/get_ticket.py"
from operations.contracts.requests.generated.preparation import GetTicketQuery
from operations.contracts.requests.generated.preparation_responses import GetTicketResponse, LoadedTicket
from operations.contexts.preparation.application.ports.ticket_read_repository import TicketReadRepository


class GetTicketQueryHandler:
    """Preserve revision, missing-resource and pagination policy in a generated response."""

    def __init__(self, read_repository: TicketReadRepository) -> None:
        self.read_repository = read_repository

    def execute(self, query: GetTicketQuery) -> GetTicketResponse:
        value = self.read_repository.get(query.id)
        if value is None:
            return {"error": {"code": "not_found", "message": ""}}
        return {
            "ticket": LoadedTicket(exists=value["exists"], version=value["version"], state=value["state"])
        }
```

### TypeScript

Source: [get-reward.ts](../../../services/engagement/src/contexts/loyalty/application/queries/get-reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/queries/get-reward.ts"
import type { GetRewardQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { GetRewardResponse } from '../../../../contracts/requests/generated/loyalty-responses.js';
import type { RewardReadRepository } from '../ports/reward-read-repository.js';
/** Preserve revision, missing-resource and pagination policy in a generated response. */
export class GetRewardQueryHandler {
  constructor(private readRepository: RewardReadRepository) {}
  async execute(query: GetRewardQuery): Promise<GetRewardResponse> {
    const value = await this.readRepository.get(query.id);
    return value ? { reward: value } : { error: { code: 'not_found', message: '' } };
  }
}
```

These handlers are intentionally small. The [read adaptor](read-repositories.md)
validates stored authority and creates the view; the query handler selects its response and the generated encoder turns
that DTO into the published reply. Keeping those roles explicit allows each
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
