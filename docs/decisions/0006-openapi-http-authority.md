# 0006 — Authoritative OpenAPI HTTP contracts

Status: accepted, 6 October 2026.

## Decision

OpenAPI 3.0.3 documents under `contracts/http/` own the HTTP methods, paths,
parameters, request objects, response objects and status codes exposed by the
Go API, Go Storefront, internal realtime gateway and development notification
provider. The API document includes every published business command and query
from all six contexts. Shared path items and schemas keep forwarded operations
and direct Storefront operations aligned.

Generate self-contained JSON bundles into each owning Go module and embed them
in its technical transport/adaptor ring. Compositions validate their document
and require every enabled operation to have a handler. They reject undeclared
registrations. The API selects its forwarding routes from the document and
retains context authentication and credential replacement in the adaptor.

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
