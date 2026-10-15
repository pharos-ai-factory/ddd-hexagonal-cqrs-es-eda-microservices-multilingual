# Implementation coverage and gaps

This report distinguishes working implementation from teaching examples. The
[concept catalogue](concepts/index.md) includes Go, Python and TypeScript snippets
on every concept page; illustrative snippets are labelled and are not wired into
the running services.

## Common implemented path

All three services implement command definition and registration, plain handler
injection, aggregate restoration and mutation, named write/read repositories,
query handlers, atomic state/outbox persistence, domain-to-integration publication,
receipts, optimistic concurrency and realtime snapshot delivery. Each concept
page links directly to the relevant source.

## Gaps in the three-language examples

| Concept                                   | Go                                        | Python                | TypeScript                    | Implication                                                                                                    |
| ----------------------------------------- | ----------------------------------------- | --------------------- | ----------------------------- | -------------------------------------------------------------------------------------------------------------- |
| Aggregate-changing durable event reaction | Implemented                               | Implemented           | Implemented                   | All three commit private command bytes before event ACK and execute through paired consumers                   |
| Event-fed consumer projection             | Implemented for Ordering's menu directory | No equivalent example | No equivalent example         | Python and TypeScript currently query owner state directly; a projection needs a specific consumer use case    |
| Standalone domain service                 | No named example                          | No named example      | No named example              | Current business rules have natural aggregate/value owners; introduce a service for a concrete new domain rule |
| Named business value-object wrapper       | Money, Quantity and menu terms            | CollectionCode        | No equivalent wrapper example | TypeScript uses validated primitive terms and read-only domain records; its Money example is illustrative      |

### Go: durable command reactions implemented

[OrderCollectedIntegrationEventHandler](../../services/storefront/contexts/ordering/application/eventhandlers/order_collected.go)
accepts `RecordCollection` through the concrete PostgreSQL command outbox.
Its owner-private codec, stored-byte fixture, paired consumers, fenced relay,
generated permissions and startup completeness checks are implemented.
[RecordCollectionCommandHandler](../../services/storefront/contexts/ordering/application/commands/record_collection.go)
loads the Order, records collection, saves and explicitly publishes its browser view.
The [recipe](concepts/durable-commands.md) links every step and recovery test.

### Python: event-fed projections

Preparation's [read repository](../../services/operations/src/operations/contexts/preparation/adaptors/persistence/ticket_read_repository.py)
maps current ticket authority to a view. Its [OrderPlaced handler](../../services/operations/src/operations/contexts/preparation/application/event_handlers/order_placed.py)
enqueues a command that creates the ticket. Neither is an independent event-fed
query projection equivalent to Ordering's menu directory.

The illustrative Python [projection example](concepts/projections.md) shows the
application shape. A real implementation also needs a consumer-owned contract,
receipt and projection transaction, subscription registration, ordering/freshness
rules and recovery tests. Engagement has the same teaching-coverage gap.

### Domain services and value objects

The [domain service page](concepts/domain-services.md) links the existing aggregate
rules in all three languages and shows a possible separate policy as a teaching
example. Its absence reflects the current business model. Adding a class solely
to fill a diagram would add no business behaviour.

Go's [Money](../../services/storefront/foundation/domain/domain.go) and Python's
[CollectionCode](../../services/operations/src/operations/contexts/collection/domain/pickup.py)
provide concrete value-object implementations. TypeScript's
[Reward](../../services/engagement/src/contexts/loyalty/domain/reward.ts) validates
plain terms. A read-only Grant with an identity is not automatically a value object.

## Deliberately outside the running reference

Event-sourced aggregate reconstruction and compensating payment/cancellation
sagas are absent in all three languages. The current implementation uses saved
aggregate state and durable forward workflows. The [event sourcing](event-sourcing.md)
and [modelling](modelling-and-boundaries.md) guides explain those concepts and
label hypothetical extensions. Production deployment is also outside the
development-only scope.

## Explicit buses and realtime publication

All three application foundations define `CommandBusPort`, `QueryBusPort`,
`DomainEventBusPort`, `IntegrationEventBusPort` and `RealtimeBusPort`. Query registries
use typed query dispatch; aggregate fact recording uses the private domain bus.
Concrete durable command adaptors implement the command bus port in all three languages.

Explicit command publication, scoped buffers, event-subscriber realtime outboxes
and typed rejection publication are implemented in all three languages. Ordering
and Preparation demonstrate domain-to-realtime subscribers. Loyalty demonstrates
an explicitly declared integration-to-realtime subscriber. The
[realtime recipe](concepts/realtime-events.md) links the source and tests.

## Generated command and query inputs

All existing public commands, private durable commands and query inputs in Go,
Python and TypeScript now derive from their Protobuf schemas. Handler modules
import the generated plain types. The generator also supplies mechanical mappings
and the shared pagination input. Owner semantic validation and all business
behaviour remain handwritten. Visibility and ownership are required schema metadata.
See the [command recipe](concepts/commands.md), [query recipe](concepts/queries.md)
and [decision 0028](../decisions/0028-schema-generated-command-and-query-inputs.md).

## Response contracts

All command outcomes and all 16 query response contracts now derive from owner
Protobuf in Go, Python and TypeScript. Generated query bindings enforce the
request/result association. Published read views alias generated resource DTOs;
owner handlers retain missing-resource, pagination and business decisions.
See [responses](concepts/responses.md) and the [query recipe](concepts/queries.md).
