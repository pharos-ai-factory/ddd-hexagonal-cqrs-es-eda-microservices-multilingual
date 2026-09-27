# 0001 — Aggregate consistency and durable delivery

Status: accepted for the initial reference, 26 September 2026.

## Decision

One command or consumed-event transaction changes at most one aggregate, along
with command receipt, outcome, consumer receipt where applicable, and outgoing
events. A projection consumer may change zero aggregates.

An immediate business invariant determines the aggregate boundary. When a
proposed operation appears to require two aggregates to commit together, revisit
the model: remove unrelated lifecycles and group the state that must be
consistent. A workflow consists of separate commands and permits intermediate
states.

The application port is bound to one context and aggregate kind. It does not
expose a transaction or an unrestricted unit of work. PostgreSQL guards reject
an accidental second aggregate write even if a caller changes the configured
target in the same transaction. This is a defence against implementation
mistakes, not a sandbox against hostile code holding a database credential.

## Domain events and integration events

Domain facts describe completed transitions without AMQP or Protobuf dependencies.
Applications select the facts that require delivery:

- Private domain contracts support reactions inside their owning context.
- Integration contracts publish facts for other contexts.

Both use the same outbox, dispatch, publisher, consumer, receipt, retry,
dead-letter and replay code. RabbitMQ topic permissions prevent another context
from binding a private domain routing key. Private contracts may evolve with their
owner, but persisted messages still require an explicit version policy.

Facts with no asynchronous reaction, such as an ordinary draft quantity change,
need not be published. There is no second, unreliable in-memory delivery path.

## Atomic persistence

Within a single PostgreSQL transaction:

1. Lock the aggregate target, including when the aggregate does not yet exist.
2. Establish consumer receipt ownership, if this is an incoming message.
3. Return a stored command outcome for a matching retry, or reject conflicting
   use of its identity.
4. Check the expected version and run the domain decision.
5. Persist the new state/version and selected immutable outgoing Protobuf bytes.
6. Record the outcome and receipts, then commit.

Expected business rejections are recorded outcomes. A database outage, invalid
persisted state or encoder failure rolls back the transaction. A retry of a
rejected command keeps its rejection; a new business attempt has a new identity.

Consumer receipt ownership is keyed independently by consumer and source event.
A concurrent conflicting copy cannot change another aggregate by selecting a
different target.

## Crash windows

| Failure | Committed evidence | Recovery |
| --- | --- | --- |
| Before producer commit | No state, receipt or outgoing event | Retry the command |
| Commit succeeds but response is lost | State, receipt and outbox together | Retry the same command identity |
| Relay crashes before publishing | Pending dispatch remains | Lease expires and another relay claims it |
| Broker confirms but dispatch completion is lost | Broker copy plus pending source dispatch | Republish; consumer receipt prevents a second effect |
| Consumer fails before database commit | No receiving transition or receipt | Retry delivery |
| Consumer commits then loses its ACK | Receiving transition and receipt | Redelivery returns the stored outcome |
| Provider accepts then its response is lost | Provider receipt under notification identity | Retry the identical provider request |
| Handler keeps failing | Original message in its dead-letter queue | Repair the cause and explicitly replay |

Delivery is at least once. Idempotency controls duplicate effects; there is no
claim of exactly-once network delivery or a distributed transaction.

Mandatory persistent publication must receive a publisher confirmation and no
unroutable return. The relay uses a token and generation to fence expired leases.
Immutable event evidence and mutable dispatch progress are separate tables.

Each consumer has durable quorum main, retry and dead-letter queues. The handler
gets one initial attempt and three delayed retries. Malformed messages go directly
to the dead-letter queue. Retry transfer uses a confirmed publication before
acknowledging the source; the retry queue uses at-least-once dead lettering.
Broker redelivery limits are disabled so they cannot silently discard a message
before the application's explicit policy acts.

## Ordering and convergence

There is no global event-order guarantee. The initial workflow uses independent
facts: preparation follows one OrderPlaced; a pickup follows one DrinksReady;
credits commute under the account lock. Published revision projections are
immutable. The account may earn a later grant before an earlier Reward has been
issued without changing either grant's entitlement.

If a future consumer requires predecessor order, add a version/gap protocol and
recovery tests. A single-consumer setting is not a substitute for that protocol.

## Provider effects

Notification delivery is outside the database transaction. The provider must
honour a stable notification idempotency key with unchanged material input.
After acceptance, a separate command records the receipt on that Notification.

The local provider simulator persists acceptance before replying and can lose a
response deliberately. This proves the protocol with that provider contract.
An arbitrary external email API may not offer that contract and cannot inherit
the guarantee automatically.

## Persistence and deployment

Current aggregate state is authoritative. The outbox retains delivered facts,
not every state transition. Replay redelivers an original message to a consumer;
it does not rebuild aggregates. Event sourcing requires a separate decision,
stream model, concurrency policy and rehydration evidence.

Each context has a separate PostgreSQL database and runtime login. A shared
server and two contexts in one service do not merge their authority. Bootstrap
owns schema and topology. All configurations and runtime guards are limited to
local/development use.
