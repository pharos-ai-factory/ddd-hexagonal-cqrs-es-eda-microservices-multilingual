# 0006 — Authoritative OpenAPI HTTP contracts

Status: accepted, 6 October 2026. API dispatch is refined by
[0007](0007-api-broker-requests.md); source layout by
[0008](0008-published-contract-ownership.md).

## Decision

OpenAPI 3.0.3 documents under `contracts/services/<service>/http_api/` own the HTTP
methods, paths,
parameters, request objects, response objects and status codes exposed by the
Go API, Go Storefront, internal realtime gateway and development notification
provider. The API document includes every published business command and query
from all six contexts. Shared path items and schemas keep API operations
and direct Storefront operations aligned.

Business source fragments live under `contracts/<context>/http_api/`: resource
paths and the owning context's schemas sit together. Service-specific fragments
and complete documents live under `contracts/services/<service>/http_api/`;
generic components live under `contracts/shared/http_api/`. The root contract
guide links to each complete document. Generated component names include source
paths to preserve distinct schemas with the same name in different owners.
Ambiguous component names fail bundling.

Generate self-contained JSON bundles into each owning Go module and embed them
in its technical transport/adaptor ring. Compositions validate their document
and require every enabled operation to have a handler. They reject undeclared
registrations. The API selects its request translation routes from the document and
retains authentication at the HTTP boundary. Internal dispatch uses the separate
Protobuf request/reply contract over RabbitMQ (0007).

Run request and response schema conformance against every Go HTTP operation in
the deterministic gate. Validate live business responses through the Go API in
the disposable integration project, including Operations and Engagement DTOs.
Negative examples establish that wrong field types, missing response fields,
unexpected properties and undeclared response statuses fail verification.

Business rules remain in their owning context. HTTP schemas describe structural
transport values; domain/application handlers decide and record expected
business rejections. Authentication, body limits, command identity, expected
versions and cursor ownership retain their existing transport checks. Schema
response validation runs in tests and the live conformance probe.

## Consequences

HTTP changes require a reviewed source-contract change and regenerated bundles.
The normal builds remain service-local and require no generator or root files.
The deterministic gate checks bundle drift; complete registration prevents
route drift when a composition starts. Live validation detects drift in the
published owner response shapes across languages.

The four bundled JSON files have generated-file exceptions to the 450-line
policy: they contain resolved path items and reusable components derived from
the smaller handwritten sources. `scripts/http_contracts.py` owns their layout.
The service-local registration helpers remain handwritten and below the limit.
