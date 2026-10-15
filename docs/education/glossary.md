# Architecture glossary

These short definitions point to the guides that explain the concepts and their
café examples. The implementation's exact rules live in the
[architecture decisions](../decisions/README.md).

| Term                     | Meaning                                                                 | Guide                                               |
| ------------------------ | ----------------------------------------------------------------------- | --------------------------------------------------- |
| Domain                   | Area of business activity being modelled                                | [Strategic DDD](strategic-ddd.md)                   |
| Subdomain                | Distinct part of the business problem                                   | [Strategic DDD](strategic-ddd.md)                   |
| Ubiquitous language      | Vocabulary shared by people and software within a model                 | [Strategic DDD](strategic-ddd.md)                   |
| Bounded context          | Boundary within which a model and its language apply                    | [Strategic DDD](strategic-ddd.md)                   |
| Context map              | Relationships and dependencies between bounded contexts                 | [Strategic DDD](strategic-ddd.md)                   |
| Published language       | Agreed interchange vocabulary between models                            | [Strategic DDD](strategic-ddd.md)                   |
| Anti-corruption layer    | Translation protecting the consumer's model from an external model      | [Strategic DDD](strategic-ddd.md)                   |
| Entity                   | Domain object with continuity of identity                               | [Tactical DDD](tactical-ddd.md)                     |
| Value object             | Domain value whose equality follows its meaningful attributes           | [Tactical DDD](tactical-ddd.md)                     |
| Aggregate                | Group of state controlled through one consistency boundary              | [Tactical DDD](tactical-ddd.md)                     |
| Aggregate root           | Entry point protecting changes to an aggregate                          | [Tactical DDD](tactical-ddd.md)                     |
| Invariant                | Business condition that must hold at the relevant transition boundary   | [Tactical DDD](tactical-ddd.md)                     |
| Repository               | Boundary for retrieving and persisting aggregates                       | [Tactical DDD](tactical-ddd.md)                     |
| Factory                  | Creation operation establishing valid initial state                     | [Tactical DDD](tactical-ddd.md)                     |
| Domain service           | Domain behaviour without a natural entity or value-object owner         | [Tactical DDD](tactical-ddd.md)                     |
| Application service      | Coordinator of a use case through domain behaviour and ports            | [Hexagonal architecture](hexagonal-architecture.md) |
| DTO                      | Data transfer object carrying a boundary's declared information         | [CQRS](cqrs.md)                                     |
| Port                     | Technology-independent application capability or requirement            | [Hexagonal architecture](hexagonal-architecture.md) |
| Adaptor                  | Translation between a port and a concrete outside technology            | [Hexagonal architecture](hexagonal-architecture.md) |
| Dependency inversion     | Source dependencies point towards policy through abstractions           | [Hexagonal architecture](hexagonal-architecture.md) |
| Dependency injection     | Dependencies are supplied to a component                                | [Hexagonal architecture](hexagonal-architecture.md) |
| Composition root         | Location that constructs and connects concrete collaborators            | [Hexagonal architecture](hexagonal-architecture.md) |
| Command                  | Request for an owner to make a decision or change state                 | [CQRS](cqrs.md)                                     |
| Query                    | Request for information without a business state transition             | [CQRS](cqrs.md)                                     |
| Read model               | Representation shaped to answer a question                              | [CQRS](cqrs.md)                                     |
| Projection               | Derived representation built from source information or events          | [CQRS](cqrs.md)                                     |
| Domain event             | Fact expressed in its owning model's language                           | [EDA](event-driven-architecture.md)                 |
| Integration event        | Published fact with an external compatibility contract                  | [EDA](event-driven-architecture.md)                 |
| Correlation              | Association of messages with a workflow                                 | [EDA](event-driven-architecture.md)                 |
| Causation                | Identification of the message that triggered subsequent work            | [EDA](event-driven-architecture.md)                 |
| Optimistic concurrency   | Checking an expected revision before accepting a write                  | [Consistency](consistency-and-delivery.md)          |
| Outbox                   | Publication intent stored atomically with local work                    | [Delivery](consistency-and-delivery.md)             |
| Receipt / inbox evidence | Durable evidence that a particular consumer accepted or processed work  | [Delivery](consistency-and-delivery.md)             |
| Idempotency              | Repeating the same operation preserves its intended effect              | [Delivery](consistency-and-delivery.md)             |
| Fencing                  | Rejecting work from an owner whose lease has been superseded            | [Delivery](consistency-and-delivery.md)             |
| Dead-letter queue        | Retained failed deliveries requiring investigation or controlled replay | [Delivery](consistency-and-delivery.md)             |
| Eventual consistency     | Related state converges through subsequent successful work              | [Consistency](consistency-and-delivery.md)          |
| Event sourcing           | Authoritative state changes are stored as an event stream               | [Event sourcing](event-sourcing.md)                 |
| Rehydration              | Restoring an aggregate from its persisted representation                | [Event sourcing](event-sourcing.md)                 |
| Snapshot                 | Saved state; its authority depends on the persistence model             | [Event sourcing](event-sourcing.md)                 |
| EventStorming            | Collaborative discovery of business behaviour through events            | [Modelling](modelling-and-boundaries.md)            |
| Choreography             | Participants coordinate by reacting to facts                            | [Workflows](modelling-and-boundaries.md)            |
| Orchestration            | A coordinator requests and tracks workflow steps                        | [Workflows](modelling-and-boundaries.md)            |
| Saga                     | Multiple local transactions with business compensation where required   | [Workflows](modelling-and-boundaries.md)            |
| BDD                      | Behaviour-driven development using shared concrete examples             | [Testing](testing-and-exercises.md)                 |

Return to [the learning path](index.md).
