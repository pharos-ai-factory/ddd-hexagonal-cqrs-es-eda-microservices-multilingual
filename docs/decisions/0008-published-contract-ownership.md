# 0008 — Published contract ownership and source layout

Status: accepted, 8 October 2026. Clarifies private delivery terminology in 0001
and shared source ownership in 0002, 0006 and 0007.

## Decision

The shared `contracts/` directory describes published interfaces. Group sources
first by bounded context, then by communication boundary:

```text
contracts/<context>/
  messaging/
    commands/v1/
    queries/v1/
    integration_events/v1/
  realtime/v1/
  http_api/
```

Create only the categories a context publishes. Query replies belong beside
their query inputs. Shared technical envelopes, metadata and HTTP components
live under `contracts/shared/`. Complete service OpenAPI documents and technical
endpoints live under `contracts/services/<service>/http_api/`.

Messaging sources describe published commands, queries and integration events. Domain facts belong to their context's domain/application code and have
no public contract version. RabbitMQ delivery does not change their ownership.
Private serialisation formats, fixtures and queue metadata live inside the
owning service's context adaptors. They support internal durable reactions and
remain implementation resources.

The shared integration envelope contains only published integration payloads.
Its removed private field numbers and names are reserved. Owner-local private
formats preserve historical field numbers and bytes so queued work, outbox
records and receipt fingerprints can recover. A local delivery format revision
is a storage compatibility concern; it creates no published version promise.

Deployment topology gathers private queue metadata from owner resources and
public bindings from the integration catalogues. Runtime descriptors include
private payloads only for their owning service. Acceptance evidence treats
foreign private payloads as opaque bytes.

PostgreSQL bootstrap SQL lives in `devops/postgres/bootstrap/`. Context migration
manifests remain with their owning persistence adaptors. These are deployment
and persistence resources. Their move preserves the existing SQL checksums.

## Consequences and validation

Developers can find a published interface by its owner and boundary. Domain
implementation remains service-local, including private asynchronous delivery.
The root contract guide links to shared resources and assembled HTTP documents.

Logical Protobuf imports and published package/message/field identities remain
stable through an explicit source map in `scripts/contract_sources.py`.
Generation remains deterministic. OpenAPI component names reflect their source
paths; expanded HTTP boundary shapes remain identical after relocation.

Native fixtures check private stored-byte compatibility and cross-language
public decoding. Architecture checks verify catalogue ownership. The complete
deterministic and infrastructure gates exercise generation, API dispatch,
outbox recovery, duplicate delivery and context permissions. Executed results
and environment limits are recorded in [verification](../verification.md).

## Service-specific generation

The [generation manifest](../../contracts/generation.md) records producer/consumer
requirements separately for events, request/reply Protobuf, realtime Protobuf and
OpenAPI. Compiler-derived local views of the shared typed envelopes retain only
declared alternatives and their transitive schema dependencies, preserving each
retained field identity and reserving omitted alternatives locally. Authoritative
published schemas and historical wire bytes remain unchanged. Browser generation
includes the API OpenAPI client and all authorised realtime contexts. Cross-language
fixtures use dedicated test outputs when a runtime has no corresponding consumer.
