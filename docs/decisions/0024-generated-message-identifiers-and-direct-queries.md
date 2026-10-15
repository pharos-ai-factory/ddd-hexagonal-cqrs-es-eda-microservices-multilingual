# 0024 — Generated message identifiers and direct query registration

Refined by [decision 0028](0028-schema-generated-command-and-query-inputs.md):
command/query payload types and mechanical input mappings are generated from
annotated owner Protobuf schemas; handlers import the generated plain contracts.

Status: accepted, 14 October 2026. Refines decisions 0014, 0017, 0019 and 0023.

## Decision

Every published command and query registration names its generated identifier,
owner input mapping and concrete handler together. Query registrations also name
the owner reply mapping. The composition root resolves the query handler with its
named read repository. The shared request registry decodes, dispatches and encodes;
command registration retains the central executor and its transaction policy.

Remove grouped `queries(singular, plural, ...)` registration and resource endpoint
wrappers. Each query now has an explicit entry in the same context composition as
its commands. Owner messaging adaptors translate Protobuf inputs into application
DTOs and map application results into replies. Query handlers retain complete
reads, explicit pagination, root revisions and missing-resource behaviour.

## Identifier sources and scope

| Identifier              | Source of truth                                   |
| ----------------------- | ------------------------------------------------- |
| Published command/query | Compiled owner Protobuf envelope payload oneofs   |
| Integration event       | Published event schema annotations                |
| Private domain delivery | Owner `private_messages.json` catalogue           |
| Private command         | Compiled owner `internal_commands.proto` envelope |
| Durable subscription    | Owner `subscriptions.json` manifest               |

`pnpm generate:contracts` emits singular, context-specific types such as
`PreparationCommand`, `PreparationQuery`, `PreparationDomainEvent`,
`OrderingIntegrationEvent`, `PreparationPrivateCommand` and
`PreparationSubscription`. Python uses `StrEnum`, TypeScript uses string enums,
and Go uses named string constants with sealed role interfaces. Members retain
the existing wire values. Services receive only their selected public families
and owner-private definitions. Private command declarations are checked against
their compiled envelope.

Published request enums live in generated request modules. Event, private-command
and subscription identifiers live in generated messaging modules. Application and
domain DTOs remain plain types. Private domain delivery identifiers describe the
transport envelope; the domain continues to queue plain facts.

Incoming event declarations, durable reactions, projection bindings and publication
subscriptions use generated identifiers. Existing typed payload-to-handler bindings,
owner validation and complete subscription checks remain in place. Generated
publication factories continue to bind published payload types to their identities.
The generated identifiers replace handwritten subscription enums and repeated
routing literals in composition and owner messaging adaptors.

## Compatibility and enforcement

This changes source-level registration APIs. Published message field numbers,
JSON names, event names, queue identities, private stored bytes and database schemas
retain their definitions. The API and browser continue to use their generated HTTP
operation contracts. Command receipts, expected versions, transaction ownership,
private-event-before-integration delivery and retry behaviour remain unchanged.

Registrations reject duplicate and wrong-context request identifiers at startup.
Static checks distinguish command, query, event and subscription identifier roles.
Architecture negative fixtures reject grouped query wrappers and literal routing
names in composition. Scaffolds, context navigation and executable examples use
the generated identifiers and direct registration APIs.

Tests cover schema-derived identifiers, context selection, private command manifest
validation, typed caller failures and direct query dispatch. Query regressions
exercise missing rows, revisions, full lists, pagination, invalid requests and
retryable storage failures. The deterministic and real-infrastructure gates verify
generated drift, compatibility, subscriptions, workflows and browser recovery.
