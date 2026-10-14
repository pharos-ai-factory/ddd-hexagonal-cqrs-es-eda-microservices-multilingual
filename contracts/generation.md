# Selective contract generation

`generation.json` declares generated contract usage separately for four categories.
Each target lists `produces` and `consumes`; an omitted service receives no output
for that category. Names identify the owning context for Protobuf and the assembled
service document for OpenAPI. Both roles select one copy of the required bindings.

| Category   | Selection and outputs                                                                                                                                                                                                                                                          |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `events`   | Context-owned integration payloads and the technical event envelope. Storefront uses Menu/Ordering; Operations uses Ordering/Preparation/Collection; Engagement uses Collection/Loyalty.                                                                                       |
| `requests` | Context-owned command/query/request/reply schema families. Each owner service receives its own families; the API receives all six. A family covers both directions: owners consume commands/queries and produce replies, while the API produces requests and consumes replies. |
| `realtime` | Context-owned Protobuf snapshots and the publication envelope. Each domain service produces its own snapshots; Web consumes all six contexts.                                                                                                                                  |
| `openapi`  | Assembled HTTP documents. API and Storefront generate their existing server/technical bundles; Web generates its HTTP operation client from the API document. API field mappings retain their Protobuf request dependencies.                                                   |

The browser's OpenAPI client serves HTTP commands and queries. Its separate
realtime Protobuf descriptor decodes Centrifugo publications. Backend integration
events and request/reply Protobuf remain outside the browser outputs.

The manifest describes generated codec requirements. An HTTP client without
schema-generated code, such as the notification provider adaptor, creates no
additional generation target. OpenAPI fragments retain their context-owned
locations; the assembled API document includes all six contexts through `$ref`.

## Schemas and dependencies

Published schemas under `contracts/<context>/` remain authoritative. The source
inventory in `scripts/contract_sources.py` preserves their stable logical import
names. Compiler descriptors supply the transitive import graph, including shared
request metadata, validation options and nested payload dependencies.

The event envelope contains technical metadata and payload bytes. Selected owned
schemas supply generated event registrations, plain DTOs and typed conversions.
The generator reads their `cafe.events.event` and `cafe.events.rules` annotations
through compiler descriptors. No handwritten event registry or payload field list
is required. Consumer-owned subscription manifests supply topology bindings.

The realtime envelope retains its typed snapshot oneof. Selection preserves each
retained field's identity and reserves omitted alternatives. Imported payload
dependencies remain available through compiler-derived dependency closure.
Runtime decoders reject unsupported event contracts or realtime payloads.

Private event/command formats stay in owner adaptors and retain their existing
generation paths. OpenAPI `$ref` dependencies, private subscriptions, topology,
replay metadata and persistence checksums retain their existing behaviour.

## Changing and checking a target

Add the schema in its owning context, and declare the producing/consuming services in the appropriate
category. New languages or OpenAPI output formats also need an explicit generator
target. Unknown services, categories, contract names and duplicate entries fail
validation. Generate with `pnpm generate:contracts`, or `pnpm generate:http` for
OpenAPI only. Review both added outputs and removed stale bindings.

`pnpm check:contracts` regenerates in isolation and detects changed, missing and
unexpected outputs. `pnpm check:compatibility` checks the complete authoritative
schemas and historical fixtures independently of local generation selections.
Selection tests preserve envelope identities and transitive dependencies, reject
invalid declarations and check browser/API category boundaries.

RewardIssued remains a runtime contract in Engagement. Its historical fixture is
also compiled and round-tripped through Go and Python in disposable test outputs;
these compatibility checks add no foreign bindings to Storefront or Operations.
The native codec tests cover each service's own event payloads, while the live
integration lane verifies the complete café workflow and browser publications.
