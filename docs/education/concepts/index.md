# Concept catalogue

Follow the [end-to-end recipe](../end-to-end-recipe.md) for the implementation order.

Each concept has its own page. Start with a definition, then follow the actual
Go, Python and TypeScript implementation through its related concepts. Code
excerpts link to their source; imports and surrounding declarations may be omitted.
The documentation check verifies that marked excerpts still match the checkout.

The catalogue below is the implementation reading order. All 31 pages are available.
The broader [learning path](../index.md) provides the introductory explanations
of DDD, hexagonal architecture, CQRS, EDA and event sourcing.

## 1. Model the business

| Concept                                 | What its page explains                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [Bounded contexts](bounded-contexts.md) | Language, ownership and the six models across three services |
| [Domain layer](domain-layer.md)         | Where business meaning and behaviour live                    |
| [Entities](entities.md)                 | Identity, child entities and root entities                   |
| [Value objects](value-objects.md)       | Validation, immutability and value equality                  |
| [Aggregates](aggregates.md)             | Consistency boundaries, mutations and pending facts          |
| [Invariants](invariants.md)             | Lifecycle rules, rejections and concurrent decisions         |
| [Factories](factories.md)               | Valid creation through an owner command                      |
| [Domain services](domain-services.md)   | When a business rule needs a separate owner                  |

## 2. Enter the application

| Concept                                         | What its page explains                                                             |
| ----------------------------------------------- | ---------------------------------------------------------------------------------- |
| [Application layer](application-layer.md)       | Use-case coordination and dependency boundaries                                    |
| [Commands](commands.md)                         | Create a callable command: DTO, schema, generation, parser, registration and tests |
| [Command handlers](command-handlers.md)         | Inject a repository, load, mutate and save                                         |
| [Queries](queries.md)                           | Define and register a read question and its reply mapping                          |
| [Query handlers](query-handlers.md)             | Inject a read repository and return a view                                         |
| [Ports](ports.md)                               | Define named application capabilities and their guarantees                         |
| [Adaptors](adaptors.md)                         | Map transport and persistence to application boundaries                            |
| [Dependency injection](dependency-injection.md) | Wire Fx, Dependency Injector and Awilix; scope repositories                        |

## 3. Restore and persist state

| Concept                                       | What its page explains                                   |
| --------------------------------------------- | -------------------------------------------------------- |
| [Write repositories](write-repositories.md)   | Owner mappings, staging and invocation lifetime          |
| [Aggregate hydration](aggregate-hydration.md) | Restore and validate current saved authority             |
| [Read models](read-models.md)                 | Shape application-owned query results                    |
| [Read repositories](read-repositories.md)     | Validate stored values, map views and preserve revisions |
| [Transactions](transactions.md)               | Commit state, outcomes, receipts and outgoing intent     |
| [Projections](projections.md)                 | Maintain derived evidence and reason about freshness     |

## 4. Deliver facts and updates

| Concept                                                     | What its page explains                                    |
| ----------------------------------------------------------- | --------------------------------------------------------- |
| [Domain events](domain-events.md)                           | Record plain facts and select private delivery            |
| [Domain event handlers](domain-event-handlers.md)           | Map committed facts or enqueue another owner command      |
| [Integration events](integration-events.md)                 | Generate public contracts and construct runtime payloads  |
| [Integration event handlers](integration-event-handlers.md) | Translate upstream facts into consumer-owned work         |
| [Durable commands](durable-commands.md)                     | Commit receiving evidence before handing off to execution |
| [Outbox](outbox.md)                                         | Persist exact outgoing bytes and recover dispatch         |
| [Receipts and idempotency](receipts-and-idempotency.md)     | Distinguish message, command and business duplicates      |
| [Realtime events](realtime-events.md)                       | Encode snapshots, commit their bytes and update browsers  |
| [Contracts and generation](contracts-and-generation.md)     | Connect owner schemas, generated code and subscriptions   |

## Follow an entire change

Read [Commands](commands.md) → [Command handlers](command-handlers.md) →
[Write repositories](write-repositories.md) → [Hydration](aggregate-hydration.md) →
[Aggregates](aggregates.md) → [Domain events](domain-events.md) →
[Transactions](transactions.md).

From the commit, follow two paths:

- **Business integration:** [Outbox](outbox.md) → [Domain event handler](domain-event-handlers.md)
  → [Integration event](integration-events.md) → [Consumer reaction](integration-event-handlers.md)
  → [Durable command](durable-commands.md).
- **Browser view:** [Realtime event](realtime-events.md) → authorised subscription
  → revision-aware view update and [query](queries.md) reconciliation.

## Reading the language examples

The recurring examples are Go's Order, Python's PreparationTicket and TypeScript's
Reward. Their use cases differ because each context owns a different decision.
The TypeScript redemption command demonstrates load/mutate/save; issuance
demonstrates creation and domain-event recording.

Aggregate-changing durable event reactions run in all three languages. Go's
`OrderCollected` → `RecordCollection` chain includes explicit browser publication.
Every concept includes Go, Python and TypeScript snippets. Missing
implementations use clearly labelled illustrative examples, alongside links to
the relevant existing source. See the [implementation coverage report](../implementation-coverage.md)
for remaining teaching gaps and the deliberate scope of the reference. Source-marked
excerpts come directly from this repository.
