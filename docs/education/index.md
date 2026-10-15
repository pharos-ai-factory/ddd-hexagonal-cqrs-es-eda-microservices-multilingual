# Learn the architecture

Start with the [end-to-end creation recipe](end-to-end-recipe.md) to build and wire
a command, its aggregate/repository, event subscribers, queries and explicit browser updates.

This learning path explains the ideas behind the café reference and shows where
they appear in working code. It assumes basic programming experience; familiarity
with DDD or distributed systems is optional. Start with the business problem,
then follow a single order through its decisions and delivery boundaries.

The café sells drinks from a published menu, prepares orders, hands them over and
rewards repeat customers. Its rules create useful modelling questions: which
prices stay fixed, who may change an order, when a collection counts, and what
happens while a reward is still being issued.

## Concept reference

Use the [concept catalogue](concepts/index.md) for one page per concept, with
Go, Python and TypeScript examples and links to the related concepts. Its 31 pages
trace command definitions and registration, handler injection, aggregate hydration
and mutation, repositories, transactions, private/public events and realtime updates.

Start at [Commands](concepts/commands.md) to follow an implementation end to end,
or use the introductory guides below to learn the broader architectural ideas.

## Learning path

| Guide                                                               | What you should be able to explain afterwards                                                 |
| ------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| [1. Strategic DDD](strategic-ddd.md)                                | Domain, subdomains, ubiquitous language, bounded contexts and context maps                    |
| [2. Tactical DDD](tactical-ddd.md)                                  | Entities, value objects, aggregates, invariants, repositories and domain services             |
| [3. Hexagonal architecture](hexagonal-architecture.md)              | Ports and adaptors, dependency direction, application services and composition                |
| [4. CQRS](cqrs.md)                                                  | Commands, queries, read models, projections and the costs of separating them                  |
| [5. Event-driven architecture](event-driven-architecture.md)        | Events, messages, contracts, domain facts and integration boundaries                          |
| [6. Consistency and reliable delivery](consistency-and-delivery.md) | Transactions, outboxes, receipts, concurrency, retries and pending workflows                  |
| [7. Event sourcing](event-sourcing.md)                              | Event streams, rehydration, snapshots and how event sourcing differs from this implementation |
| [8. Modelling and service boundaries](modelling-and-boundaries.md)  | EventStorming, modular monoliths, microservices, choreography and sagas                       |
| [9. Testing and exercises](testing-and-exercises.md)                | Which evidence proves a rule, how to trace a use case and how to reason about failures        |
| [Glossary](glossary.md)                                             | Find a concise definition and its detailed guide                                              |

## How the ideas fit together

| Concern                               | Useful concept         | Café example                                               |
| ------------------------------------- | ---------------------- | ---------------------------------------------------------- |
| Understand the business               | DDD                    | “Published menu” has an agreed meaning                     |
| Own a model                           | Bounded context        | Ordering and Preparation use different models              |
| Enforce immediate rules               | Aggregate              | Order controls all its line changes                        |
| Protect business code from technology | Hexagonal architecture | A command receives a plain repository port                 |
| Separate decisions from views         | CQRS                   | PlaceOrder changes state; GetOrder returns a view          |
| Connect independent work              | EDA                    | A published OrderPlaced fact starts Preparation's reaction |
| Survive partial failure               | Durable delivery       | A committed outbox entry remains available after a crash   |
| Choose authoritative storage          | Persistence model      | This reference saves current aggregate state               |

These concepts can be adopted independently. A small application can use an
aggregate and plain constructor ports in one process. A system can use CQRS with
one database. Event sourcing is a separate storage decision. Choose a pattern
because it solves an identified problem, and account for its operational cost.

## Reading the examples

Each guide separates a general concept from the café's implementation choices.
For example, this repository requires an aggregate-changing event reaction to
enqueue an owner command. That is a documented local convention; DDD itself
allows other transactionally safe approaches.

The [architecture](../../ARCHITECTURE.md), [domain rules](../../DDD.md) and
[accepted decisions](../decisions/README.md) define the current implementation.
The education pages explain the reasoning. The [walkthrough](../walkthrough.md)
then traces the code and running application.

You can read everything with `pnpm docs:dev`. Run the café only when you reach
the live exercises. All runtime compositions in this repository are for development.
