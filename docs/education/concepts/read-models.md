# Read models

A read model is an application-facing representation shaped for a question or
screen. It exposes the facts a reader needs while keeping domain mutation and
storage mechanics behind their own boundaries.

## The three application views

### Go

Source: [order.go](../../../services/storefront/contexts/ordering/application/readmodels/order.go).

```go source="../../../services/storefront/contexts/ordering/application/readmodels/order.go"
package readmodels

import model "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/model/ordering"

// SelectionView uses the schema-generated response representation.
type SelectionView = model.Selection

// LineView uses the schema-generated response representation.
type LineView = model.Line

// OrderView uses the schema-generated response representation.
type OrderView = model.Order
```

### Python

Source: [ticket.py](../../../services/operations/src/operations/contexts/preparation/application/read_models/ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/read_models/ticket.py"
"""Use the schema-owned response representation at the read port."""

from operations.contracts.requests.generated.preparation_responses import Ticket

TicketView = Ticket
```

### TypeScript

Source: [reward.ts](../../../services/engagement/src/contexts/loyalty/application/read-models/reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/read-models/reward.ts"
/** Schema-owned response representation used by the read repository. */
export type { Reward as RewardView } from '../../../../contracts/requests/generated/loyalty-responses.js';
```

These published views are aliases or re-exports of schema-generated plain response
DTOs. Define their fields in owner Protobuf and regenerate. They contain no
Protobuf runtime dependency. Persistence adaptors validate stored authority and
select those fields; generated encoders handle mechanical wire conversion.
Domain snapshots remain separate representations with their own invariants.
See the [response recipe](responses.md).

The read repository returns the view with its root revision. That revision lets
the browser compare a query result with a later realtime update. A view must
preserve missing-resource behaviour and complete or explicitly paginated reads.

## Direct read or maintained projection

A read adaptor can build a view directly from current saved state, as these
repositories do. An event consumer can also maintain a projection for later
queries. That introduces a separate freshness and recovery concern. A read model
describes the representation; a projection describes how a derived representation
is built or updated.

## Related concepts

[Queries](queries.md) · [Query handlers](query-handlers.md) · [Read repositories](read-repositories.md) · [Projections](projections.md) · [Realtime events](realtime-events.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
