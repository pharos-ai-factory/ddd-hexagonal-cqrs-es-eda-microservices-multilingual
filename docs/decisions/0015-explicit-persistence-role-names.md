# 0015 — Explicit persistence role names

Status: accepted, 9 October 2026. Refines the names introduced by decision 0014.

## Decision

Use `<Resource>ReadRepository` for application query persistence ports in all
three languages. They expose application read models with missing-row, version
and pagination semantics. The complete `Repository` suffix makes this role
recognisable in navigation, imports and dependency composition.

| Language | Application port file/type | PostgreSQL implementation |
| --- | --- | --- |
| Go | `ports/order_read_repository.go`: `OrderReadRepository` | `postgres/order_read_repository.go`: `postgres.NewOrderReadRepository` |
| Python | `ports/ticket_read_repository.py`: `TicketReadRepository` | `persistence/ticket_read_repository.py`: `PostgresTicketReadRepository` |
| TypeScript | `ports/account-read-repository.ts`: `AccountReadRepository` | `persistence/account-read-repository.ts`: `PostgresAccountReadRepository` |

Go factory functions use `New`. Python and TypeScript concrete classes include
`Postgres`; the Go package supplies that qualification. DI bindings and query
handler dependencies identify the read repository explicitly.

Shared `MappedReadRepository` helpers preserve query metadata while mapping
stored values. `SnapshotReadRepository` in Go and `PostgresSnapshotReadRepository`
in Python/TypeScript identify the lower-level persisted snapshot capability.
Transport query endpoints and generic query ports retain their query names: they
also represent application dispatch and HTTP/RabbitMQ boundaries.

Snapshot restoration belongs in an owner persistence module such as
`ticket_snapshot.py` or `account-snapshot.ts`, with an explicit restoration
function. Both write and read repositories use it. Read repository
modules own application view mapping. Go's existing domain restoration functions
remain the validation entry points used by the owner adaptors.

## Write repositories and central command execution

Use `<Aggregate>WriteRepository` alongside the read repository. It loads an
aggregate and saves that root through an invocation-scoped repository. Commands
can use authoritative reads on this path; visibility is governed by the primary
database, transaction isolation and version checks. A repository name alone does
not guarantee freshness.

A feature handler receives its named write repository, a narrow command context
containing the target identity, and the command. It loads the aggregate, invokes
business behaviour, saves it and returns a business status. Aggregates collect
private domain facts. Owner adaptor mappers select delivered facts and translate
them into private or published message payloads when the repository saves.
The domain never constructs transport envelopes.

The central command executor constructs a fresh handler and repository per
invocation. Infrastructure opens one local PostgreSQL transaction before loading
the target. It checks receipts and expected versions, runs the handler, then
commits the saved state, outcome, receipts, event outbox, realtime bytes and reply
bytes together. A rejection discards staged state/events and commits the rejection
outcome. Unexpected failures roll back. Incoming messages are acknowledged after
commit. Broker publication remains an independent outbox-relay operation.

The transaction coordinator is internal infrastructure. Feature handlers have no
`UnitOfWork` dependency or transaction callback. Composition uses `command.Bind`
in Go, `CommandExecutor` in Python and `bindCommand` in TypeScript. Each factory
receives its invocation's repository; no ambient or singleton mutable transaction
state is shared between commands. The notification delivery executor performs
idempotent provider I/O before acquiring the aggregate transaction. Go's
`command.BindProjection` prepares immutable projection values before that
transaction, so a feature never waits for a second pooled connection while
holding its aggregate lock.

This applies independently in each language to **one aggregate in one context's
database**. Inter-context work uses durable messages and eventual consistency.

## Alternatives and consequences

A write repository can own a transaction that saves an aggregate and its pending
events. Our command boundary additionally records rejections and no-ops where
`save` is never called, and protects loading with the existing locking policy.
Central execution covers these paths consistently without making every feature
implement them. Sharing the internal database transaction preserves atomicity.

An explicit application `UnitOfWork` interface would make coordination visible
but add callbacks and wiring to each use case. With one aggregate per command,
we keep that responsibility inside infrastructure. This still implements the
transaction-coordination behaviour described by Fowler's
[Unit of Work](https://martinfowler.com/eaaCatalog/unitOfWork.html); it does not
require a separate application abstraction.

The cost is a small executor factory and a publication mapper per aggregate.
Changes to delivered facts require updating that mapper. Named repositories,
plain handlers and central execution make the usual feature path consistent
across languages. Public contracts, SQL schemas, stored bytes and queue identities
retain their existing definitions.

Architecture guards forbid application transaction coordination and aggregate
mutation/save capabilities outside command handlers. Type, composition and BDD
checks cover the new wiring. Real PostgreSQL checks cover replay, rejection after
save, repository lifetime/identity, rollback and concurrency. Both full hand-off
gates are required.
