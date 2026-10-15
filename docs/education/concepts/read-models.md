# Read models

A read model is an application-owned representation shaped for a question or
screen. It exposes the facts a reader needs while keeping domain mutation and
storage mechanics behind their own boundaries.

## The three application views

### Go

Source: [order.go](../../../services/storefront/contexts/ordering/application/readmodels/order.go).

```go source="../../../services/storefront/contexts/ordering/application/readmodels/order.go"
package readmodels

// SelectionView is the application-owned read representation.
type SelectionView struct {
	OfferCode string `json:"offerCode"`
	Name      string `json:"name"`
	Minor     int64  `json:"minor"`
}

// LineView is the application-owned read representation.
type LineView struct {
	ID        string        `json:"id"`
	Selection SelectionView `json:"selection"`
	Quantity  int           `json:"quantity"`
}

// OrderView is the application-owned read representation.
type OrderView struct {
	ID                 string     `json:"id"`
	CustomerID         string     `json:"customerId"`
	EditionID          string     `json:"editionId"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	CollectionRecorded bool       `json:"collectionRecorded"`
	Lines              []LineView `json:"lines"`
}
```

### Python

Source: [ticket.py](../../../services/operations/src/operations/contexts/preparation/application/read_models/ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/read_models/ticket.py"
from typing import Literal, TypedDict


class TicketView(TypedDict):
    """Application read contract for one ticket, independent of storage and transport."""

    id: str
    orderId: str
    customerId: str
    instructions: str
    status: Literal["queued", "preparing", "ready"]
```

### TypeScript

Source: [reward.ts](../../../services/engagement/src/contexts/loyalty/application/read-models/reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/read-models/reward.ts"
/** Application-owned read representation of a reward. */
export type RewardView = Readonly<{
  id: string;
  grantId: string;
  customerId: string;
  benefit: string;
  status: 'issued' | 'redeemed' | 'expired';
  expiresAt: string;
  redeemedFor?: string;
}>;
```

These types belong to the application. They are separate from persisted state,
Protobuf replies and browser publication envelopes, even when their fields overlap.
Owner adaptors make the mappings explicit.

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
