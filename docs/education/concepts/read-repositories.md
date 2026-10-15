# Read repositories

A named read repository supplies application read models and their revisions.
Its interface describes read authority. Its persistence implementation validates
stored data and maps it into a stable application representation.

## Creation recipe

1. Generate the resource response DTO and alias the application view and the item/list/pagination capabilities needed by the new query.
2. Add the named `<Resource>ReadRepository` interface/protocol. Preserve resource absence and revision in the port result.
3. Implement `Postgres<Resource>ReadRepository` in owner persistence; Go uses the `postgres.New<Resource>ReadRepository` constructor. Restore/validate stored authority before selecting view fields.
4. Bind the concrete repository to the named query handler in DI. A `QueryBusPort` routes the question; this repository performs the storage read.
5. Bind the owner parser and handler through the generated response registration using the [query recipe](queries.md).
6. Test item absence, revision preservation, invalid stored state, full-list behaviour and requested pagination. Event-driven browser handlers can reuse the read port and must publish the revision actually returned with the view.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Mapping current authority into views

### Go

Source: [order_read_repository.go](../../../services/storefront/contexts/ordering/adaptors/persistence/postgres/order_read_repository.go).

```go source="../../../services/storefront/contexts/ordering/adaptors/persistence/postgres/order_read_repository.go"
package postgres

import (
	store "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/adaptors/persistence/postgres"
	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/ports"
	view "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/readmodels"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
)

// NewOrderReadRepository constructs the PostgreSQL repository for validated application views.
func NewOrderReadRepository(db *store.ContextDatabase) ports.OrderReadRepository {
	return store.MappedReadRepository[d.OrderState, view.OrderView]{Source: store.NewSnapshotReadRepository[d.OrderState](db, "order"), View: orderView}
}
func orderView(s d.OrderState) (view.OrderView, error) {
	root, err := d.Restore(s)
	if err != nil {
		return view.OrderView{}, err
	}
	s = root.Snapshot()
	lines := make([]view.LineView, 0, len(s.Lines))
	for _, l := range s.Lines {
		lines = append(lines, view.LineView{ID: l.ID, Quantity: l.Quantity, Selection: view.SelectionView{OfferCode: l.Selection.OfferCode, Name: l.Selection.Name, Minor: l.Selection.Minor}})
	}
	return view.OrderView{ID: s.ID, CustomerID: s.CustomerID, EditionID: s.EditionID, Currency: s.Currency, Status: s.Status, CollectionRecorded: s.CollectionRecorded, Lines: lines}, nil
}
```

### Python

Source: [ticket_read_repository.py](../../../services/operations/src/operations/contexts/preparation/adaptors/persistence/ticket_read_repository.py).

```python source="../../../services/operations/src/operations/contexts/preparation/adaptors/persistence/ticket_read_repository.py"
"""PostgreSQL read repository for Preparation application views."""

from operations.adaptors.persistence.postgres_database import PostgresContextDatabase
from operations.adaptors.persistence.snapshot_read_repository import PostgresSnapshotReadRepository
from operations.adaptors.persistence.mapped_read_repository import MappedReadRepository
from operations.contexts.preparation.domain.preparation_ticket import TicketSnapshot
from operations.contexts.preparation.application.read_models.ticket import TicketView
from operations.contexts.preparation.adaptors.persistence.ticket_snapshot import restore_ticket_snapshot


def ticket_view(state: TicketSnapshot) -> TicketView:
    return TicketView(
        id=state["id"],
        orderId=state["orderId"],
        customerId=state["customerId"],
        instructions=state["instructions"],
        status=state["status"],
    )


class PostgresTicketReadRepository(MappedReadRepository[TicketSnapshot, TicketView]):
    """Expose validated tickets through the Preparation read model."""

    def __init__(self, database: PostgresContextDatabase) -> None:
        super().__init__(
            PostgresSnapshotReadRepository(database, "ticket", restore_ticket_snapshot), ticket_view
        )
```

### TypeScript

Source: [reward-read-repository.ts](../../../services/engagement/src/contexts/loyalty/adaptors/persistence/reward-read-repository.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/adaptors/persistence/reward-read-repository.ts"
import type { PostgresContextDatabase } from '../../../../adaptors/persistence/postgres-database.js';
import { PostgresSnapshotReadRepository } from '../../../../adaptors/persistence/snapshot-read-repository.js';
import { MappedReadRepository } from '../../../../adaptors/persistence/mapped-read-repository.js';
import { restoreRewardSnapshot } from './reward-snapshot.js';
import type { RewardState } from '../../domain/reward.js';
import type { RewardView } from '../../application/read-models/reward.js';

/** Select the stable application read fields independently from storage. */
export const rewardView = (s: RewardState): RewardView => ({
  id: s.id,
  grantId: s.grantId,
  customerId: s.customerId,
  benefit: s.benefit,
  status: s.status,
  expiresAt: s.expiresAt,
  ...(s.redeemedFor !== undefined ? { redeemedFor: s.redeemedFor } : {}),
});

/** PostgreSQL implementation of the loyalty RewardReadRepository capability. */
export class PostgresRewardReadRepository extends MappedReadRepository<RewardState, RewardView> {
  constructor(database: PostgresContextDatabase) {
    super(new PostgresSnapshotReadRepository(database, 'reward', restoreRewardSnapshot), rewardView);
  }
}
```

These adaptors validate the stored domain representation before returning a view.
That validation may restore an aggregate internally. The query handler still
receives the read model and has no write capability.

## Preserve the contract

Item reads preserve missing roots. List reads are complete unless the caller
explicitly supplies pagination. The shared read machinery preserves revisions,
keyset continuation and the requested pagination mode. A mapping failure from
corrupt authority remains a technical error rather than a fabricated empty result.

Unlike a write repository, these pool-backed read capabilities can be supplied
to a reusable query handler at composition time. The handler does not retain an
open write transaction. See [query handler injection](query-handlers.md#inject-the-read-repository).

Inter-context reads use a published owner interface or a consumer-owned projection.
A read repository does not confer access to another context's private database.

## Related concepts

[Ports](ports.md) · [Query handlers](query-handlers.md) · [Read models](read-models.md) · [Aggregate hydration](aggregate-hydration.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
