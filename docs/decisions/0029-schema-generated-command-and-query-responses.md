# 0029 — Schema-generated command and query responses

Accepted: 15 October 2026. Extends 0028 and refines 0009, 0014, 0015 and 0024.

## Decision

Protobuf owns each bus request's input and response contract. The required
`cafe.requests.v1.contract.response` annotation names the fully qualified success
message. Public requests must select exactly one matching alternative in their
owner's `Reply.payload`, alongside the shared `RequestError`. Commands, including
private durable commands, select the shared `Outcome` produced by their executor.
Missing, foreign, unknown and incompatible associations fail generation.

Generation produces plain Go structs, Python TypedDicts and TypeScript types for
the response graph, including nested resources, repeated fields, errors and
optional presence. These types have no Protobuf runtime dependency. Generated
per-query response types constrain success and error alternatives. Generated typed
query registrations bind the request identifier, parser signature, result type
and checked wire encoder together. Go keeps registrations in a separate
`adaptors/messaging/generated/queryregistration` package so owner codecs stay independent
of the RabbitMQ worker and its persistence wiring.

The application-facing response graph lives beside generated request inputs in
service-local `contracts/requests/` modules. Existing published read-view names
become aliases or re-exports of those resource DTOs. Repeating their fields in
handwritten application modules is rejected. Internal snapshots, generic repository
`Loaded`/`Page` containers and distinct internal models retain their own roles.

## Explicit runtime responsibilities

Query handlers select missing-resource errors, successful loaded resources and
complete or explicitly paginated results. Named read repositories continue
validating stored authority and selecting published fields in owner persistence.
The generated encoder performs structural field conversion, bounds checks and
presence handling. Composition calls the generated owner registration function.
Application handlers import plain contracts and retain ordinary constructor ports.

Command handlers still return their internal business status. The central executor
assembles the generated `Outcome`, adding the committed aggregate identity/revision
or a stable rejection DTO. Domain errors remain independent of response contracts;
owner infrastructure maps their code/message into the generated rejection.
Exact API reply bytes remain part of the aggregate transaction. An encoding failure
rolls back the attempt, preserving the existing recovery and idempotency boundary.
Private commands use the same generated outcome for durable execution records.

Realtime selection remains explicit application policy. Generating a response
creates no domain, integration or browser publication.

## Schema vocabulary and compatibility

Response generation supports the current scalar types, nested message references,
repeated fields and optional scalars/messages. A message field marked
`required_output` must be present in a successful response; loaded state and Order
line selection use this rule. Optional continuations and optional resource values
preserve absence. Encoders reject integer overflow and values outside the existing
cross-language safe integer range. Maps, enums, recursive messages and user oneofs
require an explicit generator extension and fixtures.

Published response fields, their numbers, envelope identities and historical bytes
are preserved. The annotations add build-time authority and validate existing
successful-response expectations. The HTTP and browser projection schemas retain
their separate boundary definitions and explicit mappings.

Compatibility records the initial `required_output` annotation on the eight
existing loaded-state fields and Order line selection as a reviewed refinement.
The exception covers only that exact annotation; field identities, types, other
options and later removal remain checked. Declared request/response associations
are compared against historical schemas once present.

## Enforcement and evidence

Native type checks exercise generated response signatures. Negative fixtures reject
manual response copies and unbound query composition. Generator tests extend reply
schemas with nested/repeated/optional fields and check model/encoder propagation;
invalid associations and unsupported shapes fail before output is written.
Runtime tests cover nested values, optional presence, ambiguous alternatives,
numeric bounds, missing resources, revisions, pagination and infrastructure failure.

Run focused tests in every owner language, generated drift and historical
compatibility, `pnpm verify` and `pnpm test:integration`. The live lane must preserve
command reply recovery and query behaviour through RabbitMQ, the API and browser.
The query scaffold and educational recipes start with both schema directions.
