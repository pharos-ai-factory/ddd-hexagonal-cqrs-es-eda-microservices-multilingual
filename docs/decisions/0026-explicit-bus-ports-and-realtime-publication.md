# 0026 — Explicit bus ports and realtime publication

Status: accepted, 2026-10-15.

## Context

Realtime publication previously happened automatically when an aggregate snapshot
was saved. Application code could neither select its browser messages nor publish
a rejection without a saved root. `DurablePublicationPort` also obscured whether
its payload was an integration event or a browser message.

## Decision

Each context owns its realtime mappings and explicitly publishes through an
application bus port. The existing owner relays, restricted Go publisher gateway
and Centrifugo remain the delivery infrastructure. The gateway validates publisher
identity, channel and envelope; context handlers own message selection.

| Application port             | Contract                                                                         |
| ---------------------------- | -------------------------------------------------------------------------------- |
| `CommandBusPort<C>`          | Durably enqueue an owner command; acceptance precedes execution                  |
| `QueryBusPort<Q, R>`         | Dispatch a parsed application query to its handler and return a typed result     |
| `DomainEventBusPort<E>`      | Stage private aggregate facts in the owner transaction                           |
| `IntegrationEventBusPort<E>` | Commit a declared subscription receipt and exact integration bytes               |
| `RealtimeBusPort<E>`         | Explicitly stage browser messages with their owning transaction or event receipt |

`CommandBusPort` replaces `DurableCommandPort`; `IntegrationEventBusPort`
replaces `DurablePublicationPort`. Bus interfaces describe logical messaging
capabilities. Named read/write repositories retain their persistence roles.
Existing snapshot-reading ports do not become query buses.

Aggregates continue recording plain domain facts without any injected bus.
Their write adaptor maps selected facts through the private domain bus. An
independent domain-event subscriber publishes integration events, retaining the
original source revision, time and causation as required by decision 0019.

## Explicit realtime producers

Domain-event handlers, integration-event handlers and application use cases may
receive `RealtimeBusPort`. Each producer selects its message deliberately.
Integration publication carries no implicit browser-forwarding behaviour.

Commands receive an invocation-scoped bus through their registered handler factory.
Calling `publish` copies the message into a buffer. The executor closes that buffer
after invocation. The transaction checks snapshot intent against the state actually
being committed, encodes exact bytes, and commits state, receipts, outcome and
outgoing intent together. A save without publication produces no browser message.
An expected rejection discards staged snapshot updates; an explicitly selected
rejection message must match the recorded outcome. Infrastructure failures roll
back and retain retry semantics.

Event subscribers have their own manifest entries with `realtime` naming the owned
browser resource. Their bus commits the receiving receipt, exact browser bytes and
dispatch in one transaction before ACK. Redelivery verifies the original source
fingerprint and retains the existing publication. Each realtime subscription owns
one outgoing publication per source event; independently retryable outputs require
separate subscription identities. An aggregate-changing event
reaction still enqueues an owner command; it cannot announce a completed transition
before that command commits.

The executable examples use:

- Ordering: `PlaceOrder` → private `OrderPlaced` → declared realtime domain handler.
- Preparation: `CompletePreparation` → private `DrinksReady` → declared realtime domain handler.
- Loyalty: `IssueReward` → private `RewardIssued` → integration publication → declared realtime integration handler.
- Other state transitions: an explicit publication in the command handler.
- Selected rejections: Ordering quantity changes, Collection handover and Loyalty redemption.

The event-driven examples read an authoritative owner view and publish the revision
returned with that view. A delayed event can therefore refresh a newer state.
Source-event revision and snapshot revision are distinct; an old event never labels
a newer snapshot with its old revision. Missing committed resources are retryable
consistency failures. These owner-refresh handlers use owner-local identities;
foreign events need an explicit mapping to a consumer-owned resource and revision.

## Outcomes and browser behaviour

`OperationRejected` is an additive alternative in the separate realtime Protobuf
envelope. It carries command identity, correlation, stable code and safe message.
The aggregate revision can be zero when the root does not exist. Immediate request
replies and identical command retries remain the authoritative recovery path.
Timeouts continue to mean an uncertain outcome; transient failures do not become
terminal business rejection messages.

The browser validates the authorised owner/channel and payload, applies snapshots
using per-root revisions, and handles operation rejections separately. Rejection
messages are informational; they do not replace snapshot state or clear saved
uncertain command attempts. Each window deduplicates recent publication identities.
Subscriptions, history-gap reconciliation and durable session revocation retain
the existing mechanisms.

Context-local migrations allow outcome publications without an aggregate foreign
key and allow several publications at the same root revision. Receipt identity
controls producer deduplication; the frontend revision guard controls state ordering.
Database checks distinguish positive snapshot revisions from non-negative outcome
revisions. Application adaptors validate snapshot identity and committed state.

## Consequences and compatibility

This refines decisions 0001, 0002, 0011, 0015, 0019 and 0023. Atomic root-plus-realtime
commit applies when the command explicitly stages a message. An event-driven
browser update has a separate durable receiving transaction and can remain pending
after the root commits. Integration and browser updates have independent consumers,
except where a declared realtime subscription intentionally consumes integration.

Existing snapshot field numbers, private event bytes, integration identities and
command identities remain stable. New realtime subscriptions are generated into
identifiers, topology and broker permissions. Apply the owner migrations and
regenerate contracts before starting the revised development composition.

At acceptance, Go defined `CommandBusPort` without an aggregate-changing private
command example. [Decision 0027](0027-go-durable-command-reactions.md) subsequently
adds the concrete receiving outbox, paired consumers and Ordering example.
[Decision 0031](0031-targeted-realtime-messages.md) subsequently adds typed
application delivery targets and gateway-derived channels.

## Verification

Native tests cover publication scope, immutable buffers, typed query dispatch,
subscription declarations and browser decoding. PostgreSQL tests prove explicit
selection, rollback, command deduplication, event-receipt deduplication and rootless
rejection publication. Existing broker/browser lanes verify delivery recovery,
revision ordering, independent windows and session revocation. Executed results
belong in the dated verification record.
