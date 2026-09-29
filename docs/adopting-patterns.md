# Adopting these patterns in an existing system

Use the café to evaluate these patterns against an existing system's domain
model, operational needs and accepted architectural decisions. The following
questions help identify where a pattern would address a concrete problem.

| Reference mechanism | Review question or action |
| --- | --- |
| One-aggregate command port with no exposed transaction | Find handlers that synchronously mutate multiple aggregate roots; revisit their invariants and lifecycles |
| Database guard against a second aggregate target | Add mechanical evidence wherever single-aggregate ownership is intended |
| Private and public delivery through the same outbox | Replace required in-memory reactions with an atomic durable hand-off |
| Immutable events separate from dispatch leases | Keep source evidence stable while workers retry and reclaim work |
| Consumer-local receipts include bytes and target | Test conflicting identities and concurrent duplicate delivery |
| Stable recorded command outcomes | Distinguish unknown commit results from business rejections and new attempts |
| Restricted context credentials | Verify actual database and broker access, beyond naming conventions |
| Immutable published drink/menu snapshots | Prefer consumer-owned evidence when mutable owner lookups would make decisions race |
| Separate earned grant and issued Reward | Model pending workflow states instead of enlarging an aggregate to avoid eventual consistency |
| Provider idempotency plus an outcome command | Exercise acceptance followed by lost response or process failure |
| Shared HTTP journeys and Protobuf fixtures | Compare independent language services by behaviour and contracts |
| Atomic typed browser bytes plus fenced relay | Preserve the projection revision represented by a publication across retries |
| Two-window recovery tests | Prove server authorisation, recovered history and explicit reconciliation without polling |
| Optional keyset pagination | Make traversal explicit and reusable without silently truncating ordinary queries |
| Context-owned migration history | Evolve each database independently and reject mismatched runtime ownership or checksums |
| Restricted realtime gateway | Give publishers authority over their own channel and session workers only disconnection authority |
| Protected developer secret files | Use individual provider credentials with scopes and spending controls outside generated configuration |
| Architecture checker with negative fixtures | Turn enforceable architectural decisions into build failures |

## Transfer the rules with their evidence

Read each affected context's current authority and accepted architectural decisions.
Name the immediate invariant and the allowed pending state before changing a
transaction. Introduce the command/receipt/outbox boundary with rollback,
concurrency and crash-window tests in the same change.

A bounded context is larger than an aggregate, and a service is a deployment
boundary. Splitting a service does not repair a poorly chosen aggregate boundary.
Combining contexts in a service does not grant shared data access.

## Do not copy the demonstration shortcuts

The generic JSONB aggregate storage keeps this example small. Contexts own their
additional migrations and indexes. A larger system may need
typed schemas, indexes, richer query models and context-specific persistence
ports. Preserve the transaction contract rather than copying every table.

The shared `foundation/domain` contains a few teaching primitives. A real domain
may require different meanings for quantity, money and identity; do not create
a universal business model merely to share code.

The fixed operator session, CLI bearer key, pickup-code inspection, single-node
broker and local provider are development facilities. They do not replace production
authentication, authorisation, provider administration or browser subscriptions.
The current-state persistence decision also does not settle whether a particular
context would benefit from event sourcing.

Keep contributor guides, architectural decisions and executable checks consistent
when an improvement is adopted. Explain the changed rule and bring the relevant
transaction, delivery and recovery evidence into the receiving system.
