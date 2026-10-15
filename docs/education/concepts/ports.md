# Ports

A port describes an application capability or requirement in terms the owner
understands. An inbound port offers a use case. An outbound port asks a
collaborator to provide persistence, publication or another external capability.

## Message bus port names

| Port                         | Meaning                                                         |
| ---------------------------- | --------------------------------------------------------------- |
| `CommandBusPort<C>`          | Enqueue a durable owner command; acceptance precedes completion |
| `QueryBusPort<Q, R>`         | Ask a typed application question through its registered handler |
| `DomainEventBusPort<E>`      | Stage owner-private facts inside the root transaction           |
| `IntegrationEventBusPort<E>` | Commit a subscriber receipt with its exact public event bytes   |
| `RealtimeBusPort<E>`         | Explicitly publish browser snapshots or application outcomes    |

Define these interfaces/protocols in the application foundation. Implement them in
adaptors and inject narrow typed instances into handlers. Bind routing and resources
in composition. Domain aggregates queue plain facts without an injected bus.
Read/write repositories keep their persistence role names.

Sources: [Go bus contracts](../../../services/storefront/foundation/application/buses.go),
[Go integration bus](../../../services/storefront/foundation/application/publication.go),
[Python contracts](../../../services/operations/src/operations/foundation/application.py),
[TypeScript contracts](../../../services/engagement/src/foundation/application.ts).

`QueryBusPort` is used by the request registry after owner parsing. The older
`QueryPort` type represents snapshot storage reads; its `get`/`list` capability
remains distinct from message dispatch. `Bus` describes a logical capability;
several buses can share transport infrastructure.

Follow the [query recipe](queries.md), [durable-command recipe](durable-commands.md)
and [realtime recipe](realtime-events.md) to create and wire each use case.

## Named write capabilities

The owner aliases the shared technical load/save shape to its concrete aggregate.
The name makes a handler's authority clear at its constructor boundary.

### Go

Source: [order_write_repository.go](../../../services/storefront/contexts/ordering/application/ports/order_write_repository.go).

```go source="../../../services/storefront/contexts/ordering/application/ports/order_write_repository.go"
package ports

import (
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/domain"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// OrderWriteRepository loads and saves the authoritative aggregate inside its command transaction.
type OrderWriteRepository = a.WriteRepository[*d.Order]
```

### Python

Source: [ticket_write_repository.py](../../../services/operations/src/operations/contexts/preparation/application/ports/ticket_write_repository.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/ports/ticket_write_repository.py"
"""Authoritative aggregate persistence inside the preparation transaction boundary."""

from operations.foundation.write_repository import WriteRepository
from operations.contexts.preparation.domain.preparation_ticket import PreparationTicket

type TicketWriteRepository = WriteRepository[PreparationTicket]
```

### TypeScript

Source: [reward-write-repository.ts](../../../services/engagement/src/contexts/loyalty/application/ports/reward-write-repository.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/ports/reward-write-repository.ts"
import type { WriteRepository } from '../../../../foundation/write-repository.js';

import type { Reward } from '../../domain/reward.js';

/** Authoritative reward aggregate persistence inside its command transaction. */
export type RewardWriteRepository = WriteRepository<Reward>;
```

A [write repository](write-repositories.md) exposes hydrated aggregates. A
[read repository](read-repositories.md) exposes application views. Giving a query
only its read capability makes its intended role explicit.

## Ports include behaviour, not only signatures

The write port's contract includes its target, invocation lifetime, missing-state
behaviour and staging semantics. An implementation that saves a second root or
commits early violates those expectations even if its methods type-check.

The public `Execute` / `execute` operation of a handler acts as an inbound
application capability. A dedicated interface is useful when it clarifies a real
boundary; every class does not need a matching interface.

Durable command and publication ports promise an atomic hand-off with receipt
evidence. Their guarantees require database and broker tests beyond a constructor
mock. See [outbox](outbox.md) and [transactions](transactions.md).

## Related concepts

[Adaptors](adaptors.md) · [Dependency injection](dependency-injection.md) · [Write repositories](write-repositories.md) · [Read repositories](read-repositories.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
