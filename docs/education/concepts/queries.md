# Queries: define the request and response, then implement the read

A query has a schema-defined input and response. This recipe follows `GetOrder`
(Go), `GetTicket` (Python) and `GetReward` (TypeScript). Each handler uses a named
read repository and returns its generated response with the stored revision.

## 1. Name the owner and read policy

Run `pnpm context <owner>` and read its README. Identify the required fields,
missing-resource behaviour and complete or paginated list semantics. Queries have
no aggregate mutation capability. Keep private databases within their context.

## 2. Define both sides in Protobuf

Add the input in `contracts/<owner>/messaging/queries/v1/*_queries.proto` and
its response graph in `*_replies.proto`. The input's `contract.response` names the
fully qualified success message. Declare the owner and public visibility too.

### Go

```proto source="../../../contracts/ordering/messaging/queries/v1/ordering_queries.proto"
message GetOrder {
  option (cafe.requests.v1.contract) = { owner: "ordering" visibility: PUBLIC response: "cafe.ordering.requests.v1.LoadedOrder" };
  string id = 1;
}
```

```proto source="../../../contracts/ordering/messaging/queries/v1/ordering_replies.proto"
message LoadedOrder {
  bool exists = 1;
  uint64 version = 2;
  Order state = 3 [(cafe.requests.v1.required_output) = true];
}
```

### Python

```proto source="../../../contracts/preparation/messaging/queries/v1/preparation_queries.proto"
message GetTicket {
  option (cafe.requests.v1.contract) = { owner: "preparation" visibility: PUBLIC response: "cafe.preparation.requests.v1.LoadedTicket" };
  string id = 1;
}
```

```proto source="../../../contracts/preparation/messaging/queries/v1/preparation_replies.proto"
message LoadedTicket {
  bool exists = 1;
  uint64 version = 2;
  Ticket state = 3 [(cafe.requests.v1.required_output) = true];
}
```

### TypeScript

```proto source="../../../contracts/loyalty/messaging/queries/v1/loyalty_queries.proto"
message GetReward {
  option (cafe.requests.v1.contract) = { owner: "loyalty" visibility: PUBLIC response: "cafe.loyalty.requests.v1.LoadedReward" };
  string id = 1;
}
```

```proto source="../../../contracts/loyalty/messaging/queries/v1/loyalty_replies.proto"
message LoadedReward {
  bool exists = 1;
  uint64 version = 2;
  Reward state = 3 [(cafe.requests.v1.required_output) = true];
}
```

Define the resource's fields in the same reply schema. Nested resources and
repeated fields are generated recursively. `required_output` requires a successful
loaded response to supply its state. Optional fields retain their presence;
for example, an absent continuation differs from a present empty string.

## 3. Add the envelope alternatives

Add the input to `Query.payload` and its success message to `Reply.payload` in
`messaging/v1/<owner>_requests.proto`. Keep the shared `RequestError` alternative.
Allocate unused field numbers and preserve historical identities. Generation
checks that `response` identifies exactly one success alternative in this owner.

| Example   | Input member | Success member |
| --------- | ------------ | -------------- |
| GetOrder  | `get_order`  | `order`        |
| GetTicket | `get_ticket` | `ticket`       |
| GetReward | `get_reward` | `reward`       |

## 4. Generate and inspect the outputs

Run `pnpm generate:contracts`. It produces plain input/resource/response types,
wire bindings, input parsers, checked response encoders and typed registration
functions. Each query response selects its own success payload or `RequestError`.

| Language   | Plain responses                                                                                                               | Boundary bindings                                                                                                                          |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Go         | [responses.go](../../../services/storefront/contracts/requests/generated/model/ordering/responses.go)                         | [queryregistration](../../../services/storefront/contexts/ordering/adaptors/messaging/generated/queryregistration/responses.go)            |
| Python     | [preparation_responses.py](../../../services/operations/src/operations/contracts/requests/generated/preparation_responses.py) | [response_payloads.py](../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/generated/response_payloads.py) |
| TypeScript | [loyalty-responses.ts](../../../services/engagement/src/contracts/requests/generated/loyalty-responses.ts)                    | [response-payloads.ts](../../../services/engagement/src/contexts/loyalty/adaptors/messaging/generated/response-payloads.ts)                |

Plain contracts have no Protobuf runtime imports. An incompatible response handler
fails the generated binding's type signature. Commit the generated outputs.

## 5. Implement the named read repository

Use the generated resource DTO directly or through a named alias in the existing
read-model module. The named application port carries that DTO, its revision and
absence. Keep generic `Loaded`/`Page` repository containers where they express
storage results. They also serve snapshot reads and are separate from response envelopes.

Implement the port in owner persistence. Validate stored authority before selecting
published fields. Reuse the [read repository recipe](read-repositories.md), including
complete lists and explicit keyset pagination. Bind the concrete reader through DI.

## 6. Return the generated response from the handler

The handler explicitly selects `not_found` for absence and preserves the loaded
revision on success. Storage failures propagate to the registry's retryable error
handling. A successful read does not invent an aggregate transition.

### Go

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

For list queries, load every item when no page is supplied. Otherwise request the
page and preserve its continuation. Construct the generated collection DTO with
`paged` and the generated loaded resources. See
[ListOrders](../../../services/storefront/contexts/ordering/application/queries/list_orders.go),
[ListTickets](../../../services/operations/src/operations/contexts/preparation/application/queries/list_tickets.py) and
[ListRewards](../../../services/engagement/src/contexts/loyalty/application/queries/list-rewards.ts).

## 7. Retain semantic input validation

Use the generated input parser. Add UUID, authorisation or business-evidence checks
in the owner wrapper where needed. The schema supplies structural conversions.
Go's request validator checks input before the generated conversion; Python and
TypeScript wrappers retain their explicit identifier validation.

## 8. Bind the generated registration in composition

Resolve the named reader and handler before starting workers. The generated
function fixes the routing identifier, input type, response type and encoder.
Pass the owner parser and the plain handler. The shared registry dispatches through
`QueryBusPort`; generated encoders preserve nested fields, presence and numeric bounds.

### Go

```go source="../../../services/storefront/apps/storefront/di/ordering_requests.go"
responses.RegisterGetOrder(r, rpc.GetOrder, getOrder.Execute)
```

### Python

```python source="../../../services/operations/src/operations/apps/composition/preparation.py"
responses.register_get_ticket(registry, wire.get_ticket, container.get_ticket().execute)
```

### TypeScript

```ts source="../../../services/engagement/src/apps/composition/loyalty.ts"
responses.registerGetReward(registry, rewardQueries.getReward, c.getReward);
```

`rpc` imports the Go owner codecs. `responses` imports the generated owner
registration module (`adaptors/messaging/generated/queryregistration` in Go). The Go
codec package stays independent of the request worker and its persistence wiring.
The generated functions contain `wire.OrderingQueryGetOrder`,
`PreparationQuery.GET_TICKET` and `LoyaltyQuery.GET_REWARD`. Those identifiers derive
from the request envelope; the response association derives from its annotation.
There is no per-query handwritten field-copying encoder.

## 9. Expose HTTP and verify

Add the owner OpenAPI operation and the API's explicit operation-to-owner
[contract mapping](../../../services/api/adaptors/http/backend/contract_mappings.json).
Generate again, then use the generated frontend operation client. HTTP remains
its own boundary; the Go API translates the generated broker reply into it.

Test successful and missing reads, repository failures, corrupt stored authority,
nested fields, optional presence, numeric bounds and complete/paginated lists.
Registration tests exercise the actual wire route in
[Go](../../../services/storefront/apps/storefront/di/menu_queries_test.go),
[Python](../../../services/operations/tests/test_query_registration.py) and
[TypeScript](../../../services/engagement/src/apps/query-registration.test.ts).
Run the focused owner tests, `pnpm check --changed`, `pnpm verify` and
`pnpm test:integration`. Verify compatibility as well as generated drift.

Continue with [responses](responses.md), [query handlers](query-handlers.md),
[read repositories](read-repositories.md) and the [end-to-end recipe](../end-to-end-recipe.md).
