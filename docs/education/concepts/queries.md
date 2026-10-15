# Queries: create, register and expose a read use case

Use this recipe after naming the owning context and the read behaviour. The three
running examples are `GetOrder` (Go), `GetTicket` (Python) and `GetReward`
(TypeScript). Each returns an application-owned view with its stored revision.

## 1. Define the published question

Edit the owner's `contracts/<context>/messaging/queries/v1/*_queries.proto`.
Allocate unused field numbers. Define required identity, filters or pagination;
validate their business meaning in the owner input mapper.

### Go

```proto source="../../../contracts/ordering/messaging/queries/v1/ordering_queries.proto"
message GetOrder {
  string id = 1;
}
```

### Python

```proto source="../../../contracts/preparation/messaging/queries/v1/preparation_queries.proto"
message GetTicket {
  string id = 1;
}
```

### TypeScript

```proto source="../../../contracts/loyalty/messaging/queries/v1/loyalty_queries.proto"
message GetReward {
  string id = 1;
}
```

## 2. Add the query and reply to the owner envelope

Edit the owner's `messaging/v1/*_requests.proto`: add the query payload to
`Query.payload`, then add its loaded resource/list reply to `Reply.payload`.
Preserve old field numbers. A loaded response includes the resource revision.
The existing examples are:

| Language   | Owner envelope                                                                                       | Query member | Reply member |
| ---------- | ---------------------------------------------------------------------------------------------------- | ------------ | ------------ |
| Go         | [ordering_requests.proto](../../../contracts/ordering/messaging/v1/ordering_requests.proto)          | `get_order`  | `order`      |
| Python     | [preparation_requests.proto](../../../contracts/preparation/messaging/v1/preparation_requests.proto) | `get_ticket` | `ticket`     |
| TypeScript | [loyalty_requests.proto](../../../contracts/loyalty/messaging/v1/loyalty_requests.proto)             | `get_reward` | `reward`     |

## 3. Generate wire types and routing identifiers

Run `pnpm generate:contracts`. The generator reads the query envelope's oneof
members. `QueryBusPort` dispatches the resulting application query after parsing;
transport registration still uses the generated routing identifier.

| Language   | Generated identifier          | Import source                                                                                       |
| ---------- | ----------------------------- | --------------------------------------------------------------------------------------------------- |
| Go         | `names.OrderingQueryGetOrder` | [names.go](../../../services/storefront/contracts/requests/generated/cafe/requests/v1/names.go)     |
| Python     | `PreparationQuery.GET_TICKET` | [request_names.py](../../../services/operations/src/operations/adaptors/generated/request_names.py) |
| TypeScript | `LoyaltyQuery.GET_REWARD`     | [request-names.ts](../../../services/engagement/src/adaptors/generated/request-names.ts)            |

`names` is a Go import alias. `PreparationQuery` and `LoyaltyQuery` are generated
routing enums. Creating an application query class does not add an envelope member
or a routing identifier; these are separate steps.

## 4. Create the input and named query handler

Create `application/queries/<question>` in the owner. Keep the DTO and handler
together. Inject the named read repository; return its application view, revision
and missing-resource result. These complete files show the boundary:

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

## 5. Define the view and implement its read repository

Add the required fields to an application read model, then define the named port.
Implement it in the context's persistence adaptor. Validate stored state before
mapping it. The existing implementations pair these files:

| Language   | Read model                                                                                                      | Application port                                                                                                                          | PostgreSQL implementation                                                                                                                    |
| ---------- | --------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Go         | [order.go](../../../services/storefront/contexts/ordering/application/readmodels/order.go)                      | [order_read_repository.go](../../../services/storefront/contexts/ordering/application/ports/order_read_repository.go)                     | [order_read_repository.go](../../../services/storefront/contexts/ordering/adaptors/persistence/postgres/order_read_repository.go)            |
| Python     | [ticket.py](../../../services/operations/src/operations/contexts/preparation/application/read_models/ticket.py) | [ticket_read_repository.py](../../../services/operations/src/operations/contexts/preparation/application/ports/ticket_read_repository.py) | [ticket_read_repository.py](../../../services/operations/src/operations/contexts/preparation/adaptors/persistence/ticket_read_repository.py) |
| TypeScript | [reward.ts](../../../services/engagement/src/contexts/loyalty/application/read-models/reward.ts)                | [reward-read-repository.ts](../../../services/engagement/src/contexts/loyalty/application/ports/reward-read-repository.ts)                | [reward-read-repository.ts](../../../services/engagement/src/contexts/loyalty/adaptors/persistence/reward-read-repository.ts)                |

For an item query, preserve absence explicitly. For a list query, return every item
unless pagination is requested; then preserve the continuation. See the
[read repository recipe](read-repositories.md).

## 6. Translate the generated request into the application DTO

Add the mapper to the owner's messaging adaptor. These are handwritten functions;
the names `rpc`, `wire` and `rewardQueries` are import aliases for these modules.
They decode generated transport messages into the inputs from step 4.

### Go

Source: [order_queries.go](../../../services/storefront/contexts/ordering/adaptors/messaging/order_queries.go).

```go source="../../../services/storefront/contexts/ordering/adaptors/messaging/order_queries.go"
func GetOrder(value *pb.GetOrder) q.GetOrderQuery { return q.GetOrderQuery{ID: value.GetId()} }
```

### Python

Source: [ticket_requests.py](../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py).

```python source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py"
def get_ticket(value: Message) -> GetTicketQuery:
    """Decode the published item query into its application input."""
    if not isinstance(value, GetTicket):
        raise ValueError("Unexpected get ticket payload")
    return GetTicketQuery(identifier(value.id))
```

### TypeScript

Source: [reward-query-mapping.ts](../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-query-mapping.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-query-mapping.ts"
export function getReward(value: unknown): GetRewardQuery {
  return { id: identifier(text(record(value).id)) };
}

/** Preserve complete reads and explicitly requested pagination. */
```

## 7. Map the result into its published reply

Implement `GetOrderReply`, `get_reply` or `getReply` in the same owner mapper.
A missing resource becomes the existing `not_found` reply. A present view includes
its revision. Keep this mapping outside the query handler.

### Go

Source: [order_queries.go](../../../services/storefront/contexts/ordering/adaptors/messaging/order_queries.go).

```go source="../../../services/storefront/contexts/ordering/adaptors/messaging/order_queries.go"
func GetOrderReply(_ q.GetOrderQuery, loaded a.Loaded[view.OrderView]) envelope.Reply {
	if !loaded.Exists {
		return &pb.Reply{Payload: &pb.Reply_Error{Error: &envelope.RequestError{Code: "not_found"}}}
	}
	return &pb.Reply{Payload: &pb.Reply_Order{Order: orderReply(loaded).(*pb.LoadedOrder)}}
}
```

### Python

Source: [ticket_requests.py](../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py).

```python source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py"
def get_reply(query: GetTicketQuery, loaded: Loaded[TicketView] | None) -> Reply:
    """Map a missing resource or its current revision to the owner reply."""
    if loaded is None:
        reply = Reply()
        reply.error.code = "not_found"
        return reply
    return Reply(ticket=item(loaded))
```

### TypeScript

Source: [reward-query-mapping.ts](../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-query-mapping.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-query-mapping.ts"
export function getReply(_query: GetRewardQuery, loaded: Loaded<RewardView> | undefined) {
  return loaded ? { reward: reward(loaded) } : { error: { code: 'not_found' } };
}
```

## 8. Resolve the repository and handler through DI

Bind the concrete PostgreSQL read repository to its named application port, then
construct the query handler with that repository before starting workers.

- Go: [ordering.go](../../../services/storefront/apps/storefront/di/ordering.go) binds `ports.OrderReadRepository` and constructs `GetOrderQueryHandler`.
- Python: [preparation.py](../../../services/operations/src/operations/apps/composition/preparation.py) binds `ticket_read_repository` and `get_ticket` providers.
- TypeScript: [loyalty.ts](../../../services/engagement/src/apps/composition/loyalty.ts) binds `rewardReadRepository` and `getReward` factories.

## 9. Register identifier, parser, handler and reply mapper together

The registry builds the typed `QueryBusPort` dispatch. Each query has one handler;
its repository remains a distinct persistence dependency.

### Go

```go source="../../../services/storefront/apps/storefront/di/ordering_requests.go"
	requests.Query(r, names.OrderingQueryGetOrder, rpc.GetOrder, getOrder.Execute, rpc.GetOrderReply)
```

### Python

```python source="../../../services/operations/src/operations/apps/composition/preparation.py"
    registry.query(
        PreparationQuery.GET_TICKET, wire.get_ticket, container.get_ticket().execute, wire.get_reply
    )
```

### TypeScript

```ts source="../../../services/engagement/src/apps/composition/loyalty.ts"
registry.query(LoyaltyQuery.GET_REWARD, rewardQueries.getReward, c.getReward, rewardQueries.getReply);
```

## 10. Expose HTTP and verify the entire route

Add the operation to the owner's OpenAPI fragments under `http_api/`. Add its
explicit operation-to-owner request translation to the API's
[contract mappings](../../../services/api/adaptors/http/backend/contract_mappings.json).
Run `pnpm generate:contracts`, then call the generated frontend operation client.
The API sends the generated Protobuf query through RabbitMQ to this registration.

Test successful reads, missing resources, corrupt state, repository failure,
reply revision, complete lists and requested pagination. Use these registration
examples: [Go](../../../services/storefront/apps/storefront/di/menu_queries_test.go),
[Python](../../../services/operations/tests/test_query_registration.py),
[TypeScript](../../../services/engagement/src/apps/query-registration.test.ts).
Run the owning `pnpm test:focused <service> --context <context>` lane, followed by
`pnpm check --changed`, `pnpm verify` and `pnpm test:integration`.

Continue with [query handlers](query-handlers.md), [read repositories](read-repositories.md),
[realtime reconciliation](realtime-events.md) and the [end-to-end recipe](../end-to-end-recipe.md).
