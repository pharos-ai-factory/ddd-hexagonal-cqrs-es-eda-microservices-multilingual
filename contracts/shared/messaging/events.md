# Integration event contracts

Each producing context owns its published Protobuf messages under
`contracts/<context>/messaging/integration_events/v1/`. A message's
`cafe.events.event` option declares its name, owner, aggregate kind and version.
`cafe.events.rules` declares portable field constraints. Those definitions generate
catalogues, plain DTOs, typed conversions and publication factories.

The [technical envelope](v1/events.proto) carries event identity, name, owner,
version, source aggregate metadata, correlation, causation, occurrence time and
encoded payload bytes. It imports no business schemas. Generated service-local
registrations select the payload type. The receiver checks ownership, version,
identities and schema constraints before exposing plain data to the application.

## Add an event

1. Add a message to an owned schema, or create a new `.proto` file in its existing
   event directory. Include the event metadata and field constraints. New files
   need the same package and Go package options as their context's other schemas.
2. Queue an explicit `*DomainEvent` on the aggregate. Add its owner-private schema,
   codec and metadata, then stage that private fact through the repository mapper.
   Define an independent `*DomainEventHandler` using `DurablePublicationPort` to
   map the delivered fact to the generated `*IntegrationEvent` DTO. Bind the port's
   adaptor to the generated publication factory and declare `publication` in the
   private subscriber's manifest.
3. Each receiver declares its own `adaptors/messaging/subscriptions.json` entry
   and binds its projection or durable-command reaction in composition. Declare a
   new contract-family dependency in `contracts/generation.json` when needed.
4. Run `pnpm generate:contracts`, inspect the generated changes, and test the
   contract, semantic mapping and observable business outcome.

A new ordinary event requires no handwritten shared-codec or registry changes.
An event can have zero or many independent consumers. Producer schemas contain
no consumer names. Deployment topology joins subscription declarations to the
published contracts and owner-private delivery metadata.

## Type and validation rules

Generated published DTOs use explicit `IntegrationEvent` suffixes and plain
language types with no Protobuf dependency. Published Protobuf names and wire
identities retain their existing spelling for compatibility.
Application commands, domain facts and read models retain their own definitions.
The generator supports strings, booleans, int32/int64, uint32/uint64, singular
messages and repeated messages. Integers must fit JavaScript's safe-integer range
as well as their wire type. Protobuf scalar defaults are retained; repeated
messages become complete arrays. Required identities are explicit UUID rules.
Unsupported optional presence, enums, maps, scalar arrays, nested declarations and
recursive payloads fail generation and need a reviewed extension.

Constraints cover UUIDs, source identity, non-empty strings, timestamps with
explicit offsets, portable regex patterns, numeric bounds and collection bounds.
Context adaptors retain semantic checks involving multiple fields. Aggregates
own the business rules governing state transitions.

## Delivery

AMQP uses `cafe.events`, routing key `integration.<event-name>`, persistent
`application/x-protobuf` messages and publisher confirms. Broker identity, type,
owner, correlation and contract version agree with the envelope. Each receiving
subscription has its own receipt; the original bytes supply its fingerprint.

Private domain facts remain plain owner types. Private wire schemas, codecs and
metadata live in context adaptors. Private subscriptions stay inside that owner.
Both private and published deliveries use the transactional outbox, manual ACK,
bounded retry, dead-letter and replay path. Aggregate-changing reactions durably
enqueue an owner command before ACK; projection reactions update consumer reads.
Publication reactions commit the subscription receipt, exact integration bytes and
dispatch together before ACK. Broker publication and publisher confirm happen in
the relay afterwards. The source revision, occurrence time and correlation survive
the mapping; the private event ID becomes causation. See
[decision 0019](../../../docs/decisions/0019-durable-domain-event-publication.md).

The public envelope was redesigned by [decision 0017](../../../docs/decisions/0017-generated-event-boundaries.md).
The shared RewardIssued fixture now uses the generic envelope. Private formats,
API request contracts and realtime projections retain their separate roles.
Run `pnpm verify` and `pnpm test:integration` before hand-off.
