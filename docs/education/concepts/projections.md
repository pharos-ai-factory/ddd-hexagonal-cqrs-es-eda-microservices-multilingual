# Projections

A projection builds or updates a derived representation from source facts.
It can supply a query view or evidence needed by a later command. Its owner must
define freshness, identity, duplicate handling and recovery.

## Go: Ordering's menu directory

Ordering consumes the published menu so that it can use immutable offer terms
without reading Menu's database. Each published edition has its own identity.

### Go

Source: [menu_published.go](../../../services/storefront/contexts/ordering/application/projections/menu_published.go).

```go source="../../../services/storefront/contexts/ordering/application/projections/menu_published.go"
package projections

import (
	"context"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/generated/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// MenuDirectoryProjectionHandler updates the consumer-owned projection without aggregate mutation.
type MenuDirectoryProjectionHandler struct {
	Directory a.ProjectionPort[model.MenuPublishedIntegrationEvent]
}

func (h MenuDirectoryProjectionHandler) Handle(ctx context.Context, m a.Metadata, event model.MenuPublishedIntegrationEvent) (a.Outcome, error) {
	// Each edition is published once; later menus have a new edition identity.
	return a.Outcome{}, h.Directory.Record(ctx, m, event.EditionID, 1, event)
}
```

The projection port records receiving evidence and the derived value. It does
not mutate an Order. Creating or changing an Order remains a separate command.

## Python and TypeScript: read views and event reactions

The current Operations and Engagement examples use direct read mappings for
their ticket and reward views. Their aggregate-changing event reactions enqueue
commands. They have no equivalent menu-directory projection handler to copy.
The three language [read repository examples](read-repositories.md) show those
direct views, while [integration event handlers](integration-event-handlers.md)
shows the actual Python and TypeScript reactions.

### Python: illustrative projection handler

This small equivalent is a teaching example. Its input is deliberately reduced
to edition identity; a real consumer would use the generated published-menu
contract and retain its offer terms. The persistence port must atomically record
the incoming receipt and projected value.

```python
from typing import Protocol, TypedDict


class PublishedMenu(TypedDict):
    edition_id: str


class MenuProjectionPort(Protocol):
    def record(self, event_id: str, edition_id: str, menu: PublishedMenu) -> None: ...


class MenuDirectoryProjectionHandler:
    def __init__(self, directory: MenuProjectionPort) -> None:
        self.directory = directory

    def handle(self, event_id: str, menu: PublishedMenu) -> None:
        self.directory.record(event_id, menu["edition_id"], menu)
```

Operations currently has [direct ticket read mapping](../../../services/operations/src/operations/contexts/preparation/adaptors/persistence/ticket_read_repository.py).
The projection port, subscription and persistence implementation above are absent.

### TypeScript: illustrative projection handler

The corresponding TypeScript example has the same receipt/derived-state contract:

```ts
type PublishedMenu = Readonly<{ editionId: string }>;

interface MenuProjectionPort {
  record(eventId: string, editionId: string, menu: PublishedMenu): Promise<void>;
}

class MenuDirectoryProjectionHandler {
  constructor(private directory: MenuProjectionPort) {}

  handle(eventId: string, menu: PublishedMenu): Promise<void> {
    return this.directory.record(eventId, menu.editionId, menu);
  }
}
```

Engagement currently has [direct reward read mapping](../../../services/engagement/src/contexts/loyalty/adaptors/persistence/reward-read-repository.ts).
It has no event-fed menu-directory projection. Both examples require a real
transactional adaptor and subscription before they could be used. The
[coverage report](../implementation-coverage.md) records these gaps.

## Freshness is part of behaviour

Ordering can reject a request as `menu_pending` while its projection is missing.
That recorded rejection remains the outcome of an identical command retry.
After the projection arrives, a new business attempt uses a fresh command ID.
The [recovery scenarios](../../../specifications/ordering/projection-recovery.feature)
make the distinction executable.

A revision can prevent an older update from replacing newer state where the
contract defines such an ordering. A timestamp alone does not establish a total
order across independent publishers.

## Rebuilding needs a sufficient source

A projection can be rebuilt only if an authoritative source retains enough
information. This reference's outbox is a delivery mechanism and does not promise
a complete historical event store. Do not assume all projections can be rebuilt
from retained deliveries. See [event sourcing](../event-sourcing.md).

## Related concepts

[Read models](read-models.md) · [Read repositories](read-repositories.md) · [Integration event handlers](integration-event-handlers.md) · [Receipts and idempotency](receipts-and-idempotency.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
