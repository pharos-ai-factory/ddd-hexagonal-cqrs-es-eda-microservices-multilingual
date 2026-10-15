# 0027: Go durable command reactions

Status: accepted. Date: 15 October 2026.
Refines decisions 0011, 0023, 0024 and 0026.

## Context

Python and TypeScript already accept event-triggered aggregate commands durably.
Go defined `CommandBusPort` but lacked a concrete private command path. Educational
recipes therefore could not demonstrate this boundary end to end in every language.

## Decision

Implement the same two-transaction receiving policy in Go. A typed owner codec
encodes immutable private Protobuf bytes and preserves the original source receipt.
`CommandOutbox` commits the receiving record and dispatch intent before event ACK.
A fenced relay publishes persistent, mandatory bytes to the owner's command queue
and completes dispatch only after broker confirmation. The private command consumer
preserves the accepted identity and runs the existing central aggregate executor.
Aggregate state, execution receipts, outcome and explicit realtime intent commit
together before command ACK. Infrastructure errors remain retryable; recorded
business rejections enter the dead-letter path. Recovery preserves original bytes.

`registration.Subscribe` binds the typed private codec, source-event target mapper,
reaction and plain command-handler factory together. It constructs the executor,
registers both consumers and supplies the relay decoder. Startup validates the
complete manifest, including command consumers and decoders. Both domain and
integration source events can use this helper. Generated topology and permissions
restrict each queue family to its owning credential. Application handlers receive
plain ports and never resolve DI or invoke another command handler.

## Ordering example

`OrderCollectedIntegrationEventHandler` maps Collection's fact to
`RecordCollectionCommand`, targeting the payload's Order ID. The source aggregate
is a Pickup. The command loads the Order through `OrderWriteRepository`, requires
placement and matching customer, calls `RecordCollection` and saves a changed root.
Missing order authority is an infrastructure/consistency failure requiring recovery.
A draft order or mismatched customer is a typed business rejection.

Collection retains authority over physical handover. Ordering records its observation
in `collectionRecorded`; placement status remains `placed`, and lines stay frozen.
A collected Pickup awaiting Ordering's command is an allowed pending state. Source
receipt identity deduplicates delivery. The aggregate's recorded flag prevents a
second transition when another source event ID describes the same collection.

The aggregate queues `OrderCollectionRecordedDomainEvent` locally. No delivery or
integration mapping is declared for that fact. The command explicitly publishes
the updated Order snapshot through `RealtimeBusPort`, atomically with its save.
Repeated collection facts produce no new revision or browser publication. Other
subscribers to `OrderCollected` remain independent.

## Compatibility and administration

Append context migration 0004 to Menu and Ordering. Both Go owner databases now
have the standard command acceptance/dispatch tables and diagnostic backlog.
Only the administrative migration runner applies DDL. Runtime privileges permit
acceptance inserts and dispatch progress; accepted bytes cannot be updated.

Add Collection to Storefront's event selection and declare
`ordering.record-collection` with private command `recordCollection`. Generation
adds its private command names, typed Collection DTO, Protobuf bindings, paired
queue catalogue and exact broker permissions. Preserve all existing wire identities.
The owner-private command envelope starts with a deterministic compatibility fixture.

Order query and realtime Protobuf messages add boolean field 7. OpenAPI adds the
optional `collectionRecorded` field. Historical snapshots default to false and
existing placement status values remain unchanged. The compatibility checker compares
the disjoint array/page-object response branches recursively and retains strict
checks for other unions, removed guarantees, field types and enum expansion. The browser displays the recorded
collection from Ordering's own view without issuing additional business queries.

## Verification

Native tests cover the invariant, source receipt preservation, historical private
bytes, invalid envelopes and complete worker registration. PostgreSQL tests cover
acceptance rollback, concurrent duplicates, conflicting evidence, immutable bytes,
fenced recovery and atomic aggregate/receipt/realtime execution. RabbitMQ tests
exercise private command retries, dead letters and explicit replay. The live journey
pauses the Go command consumer after acceptance, restarts it and observes completion.
The browser verifies the resulting Order update in two independent windows.

See [durable command recipes](../education/concepts/durable-commands.md) and
[executed evidence](../verification.md).
