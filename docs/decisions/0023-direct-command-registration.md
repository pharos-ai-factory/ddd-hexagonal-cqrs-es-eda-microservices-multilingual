# 0023 — Direct command-to-handler registration

Status: accepted, 14 October 2026. Refines decisions 0010, 0015 and 0016.

[Decision 0024](0024-generated-message-identifiers-and-direct-queries.md) extends
registration to queries and generates command, query, event and subscription identifiers.

Realtime publication and bus naming are refined by [decision 0026](0026-explicit-bus-ports-and-realtime-publication.md).
Browser messages are explicitly selected; event subscribers commit their output
with a receiving receipt, and command publications commit with the command outcome.

## Decision

Feature composition declares the incoming command name or private codec, its
owner parser and its plain handler factory together. Configure the transaction
once per aggregate and use the shared command registration API for each command.
The helper constructs the central executor internally. Per-command executor
providers and transport registrations spread across separate modules are removed.

| Language   | Shared registration helper     | Feature declarations               |
| ---------- | ------------------------------ | ---------------------------------- |
| Go         | `apps/commandregistration`     | `apps/storefront/di/*_requests.go` |
| Python     | `apps/command_registration.py` | `apps/composition/<context>.py`    |
| TypeScript | `apps/command-registration.ts` | `apps/composition/<context>.ts`    |

Python and TypeScript expose `request` and `subscription` methods on an
aggregate's `CommandRegistration`. Go uses `registration.Register` and
`registration.RegisterProjection` with an aggregate registry. All APIs preserve
static command/handler compatibility. Handler factories receive a fresh repository
inside the transaction. Extra dependencies are resolved before workers start and
captured by the factory; execution never resolves the DI container.

Published payload conversion stays in the owning context's messaging adaptor.
The service root starts the completed request registries. Private subscription
registrations name their codec and concrete handler factory together, retain the
paired event/command consumers and validate the complete manifest at startup.

Preparation keeps its AcceptOrder, StartPreparation and CompletePreparation
handlers. Collection keeps OpenPickup and CollectOrder. Menu's Drink and
MenuEdition, Ordering's Order, Loyalty's Account and Reward, and Communication's
Notification retain their existing business rules and allowed pending states.
Queries and publication reactions keep their existing registrations.

## Preparation before transactions

Go's projection registration reads the immutable published projection before
opening the aggregate transaction. This preserves the single-connection execution
path and avoids waiting for a second pooled connection under an aggregate lock.

Communication uses `NotificationDeliveryPreparation` to validate stored content
and obtain an idempotent provider receipt. `preparedSubscription` passes that
invocation's receipt to `DeliverNotificationCommandHandler` after opening the
transaction. Provider failures propagate to the command retry mechanism before
any aggregate transaction starts. Each concurrent delivery has its own receipt.

## Compatibility and consequences

Central command execution continues to own receipts, expected versions, rejection
outcomes, atomic state/outbox/realtime/reply persistence and repository lifetime.
Command IDs, source receipt material, private codecs, queues, public contracts
and database schemas retain their existing definitions. Loyalty preserves its
derived collection/grant command identities. Communication preserves the mapped
notification command as its receipt material.

Developers can follow each command directly to its concrete handler in one
registration. The shared helper adds a small composition abstraction and keeps
execution policy available for infrastructure tests. Fx, Dependency Injector and
Awilix continue to own resources and dependency graphs.

Architecture checks reject executor construction and lower-level command routing
in feature composition. Negative fixtures cover all three languages. Composition
tests dispatch real handlers through the registrations; infrastructure tests
continue to prove transaction, reply and delivery recovery. Scaffolds and the
[developer workflow](../developer-workflow.md) use the same registration APIs.
