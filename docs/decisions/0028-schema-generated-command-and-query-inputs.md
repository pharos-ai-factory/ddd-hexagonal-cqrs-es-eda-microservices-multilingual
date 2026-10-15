# 0028 — Schema-generated command and query inputs

Accepted: 15 October 2026. Refines decisions 0013, 0014, 0017, 0024 and 0027.

## Decision

Every bus command and query has an authoritative Protobuf payload. Each payload
explicitly declares its owning context and `PUBLIC` or `PRIVATE` visibility through
`cafe.requests.v1.contract`. Generation checks that annotation against its owner
and delivery envelope. Missing annotations, foreign ownership, unspecified
visibility, duplicate inputs and unbound annotated payloads fail generation.

Public commands and queries describe supported external capabilities. Private
commands describe owner-controlled durable workflow steps. Both produce plain
application-facing types through the same descriptor-driven generation pipeline.
Public schemas retain their versioned published locations; private schemas retain
their owner-local locations and historical queued-byte compatibility.

Generation supplies Go structs, Python command `TypedDict`s and query dataclasses,
TypeScript plain types, mechanical input mappings, and the existing typed routing
identifiers. Plain types live under each service's `contracts/requests/` tree and
import no Protobuf runtime. The shared pagination input is also generated.
Handlers import these types in their existing business-named application files.
Application code retains semantic dependence on its contract and independence
from messaging infrastructure. Domain models remain independent of request contracts.

One handwritten handler belongs in each command/query file. Repeating the payload
fields in that file is rejected by the native architecture checks. An intentional
translation to a different internal operation needs an explicit architectural
review; the current use cases all use their generated schema inputs directly.

## Mapping and validation

Go owner adaptors receive already validated public wire messages and use generated
field conversions. Existing request validation checks required presence and numeric
bounds before dispatch. Private Go codecs use generated payload conversions and
retain owner identity, receipt and envelope checks.

Python and TypeScript generated parsers narrow scalar values and preserve optional
presence. Owner wrappers retain semantic checks such as UUID validity and grant
terms. Their private codecs continue serialising exact Protobuf payloads through
the shared codec. Integration-to-command mapping remains an explicit owner policy.

Required public scalar fields retain explicit Protobuf presence and `required_input`.
A supplied zero differs from an omitted required input. Optional scalars and message
fields retain their absence. An absent page requests the complete list; a present
page reaches the named read repository with its limit and optional cursor.

The current generator supports the scalar request vocabulary and the shared
PageRequest message. Repeated fields, enums, arbitrary nested messages and user
oneofs require an explicit generator extension and negative/round-trip fixtures.
Unsupported shapes fail generation. Existing wire field numbers, names and
required-input annotations retain their meaning.

## Visibility and broker authority

Public request envelopes and private durable-command envelopes have different
delivery protocols. Visibility must agree with that protocol; changing an
annotation alone cannot expose private work through a public request envelope.
The schema catalogue validates durable subscription command references and feeds
the generated broker permission catalogue together with API contract selection.

Public request routing remains scoped to owner and command/query kind. RabbitMQ
checks routing keys; individual payload validation occurs at the owner request
boundary. Private queues and delivery exchanges retain owner-only permissions.
Visibility grants no universal publishing permission. A new caller or a change
of visibility requires the corresponding supported envelope, caller selection,
registration and compatibility review.

## Compatibility and verification

This changes source imports and generator responsibilities. It preserves all
existing payload wire identities, version envelopes, durable fixture bytes,
transaction boundaries, handler behaviour and explicit realtime selection.
Schema options add build-time metadata; payload encodings retain their fields.

Verification includes schema-field evolution, visibility/owner/shape failures,
generation drift, native handler imports, query presence and numeric edge cases,
existing registration and behavioural suites, historical compatibility and the
real PostgreSQL/RabbitMQ/browser workflow lane. The command/query recipes begin
with the schema and show generated imports, mappings, DI and registration.
