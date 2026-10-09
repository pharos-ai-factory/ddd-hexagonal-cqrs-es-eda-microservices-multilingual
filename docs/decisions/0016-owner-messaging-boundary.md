# 0016 — Owner business requests enter through messaging

Status: accepted, 9 October 2026. Supersedes the development inspection HTTP
exception in [0007](0007-api-broker-requests.md). Refines the HTTP surface in
[0006](0006-openapi-http-authority.md). Command execution continues to follow
[0015](0015-explicit-persistence-role-names.md).

## Decision

Expose business HTTP commands and queries through the Go API. It validates the
OpenAPI request, maps it to the owner's versioned Protobuf message and sends it
through RabbitMQ. The owner's messaging adaptor maps that message to a plain
application command or query. Composition injects typed command executors and
query handlers directly into the messaging registry.

Remove context `adaptors/http/` modules and owner HTTP command/query helpers in
Storefront, Operations and Engagement. Each owner process serves only health
and authenticated diagnostics over HTTP. Centrifugo callbacks, the realtime
gateway and notification-provider I/O retain their dedicated technical transports.

OpenAPI business fragments remain context-owned under
`contracts/<context>/http_api/`: they describe the public API's business surface.
Storefront's service OpenAPI document now describes health and diagnostics.
Application handlers continue to receive repository ports; transport and
transaction details remain in adaptors and composition.

## Rationale and trade-offs

A single runtime request path makes authentication, mapping, command identity,
expected versions, durable replies and failure recovery consistent. Developers
can trace a feature from its public HTTP operation through its owner messaging
adaptor to its command/query module. Removing the duplicate JSON decoders and
route registrations also reduces maintenance and contract drift.

Local HTTP inspection now uses the API and requires RabbitMQ. A broker outage
therefore prevents those business reads as well as commands. Focused application
tests still invoke handlers with repository probes, and authenticated diagnostics
remain available on each owner process. This accepts the extra local dependency
so development exercises the same transport boundary as the full composition.

## Compatibility and enforcement

The 17 direct Storefront business operations are deliberately retired. The
compatibility gate contains a fixed, tested list of those exact method/path
removals in `scripts/http_retirements.py`. API operations, authentication,
operational HTTP and Protobuf/stored-message contracts retain their compatibility
checks. Operations and Engagement also lose their direct development business
routes; those routes had no separate published service OpenAPI documents.

`pnpm check:architecture` rejects context HTTP adaptors and business route
literals in shared owner HTTP/composition code. Native tests verify the operational
surface and reject every public business path with authenticated owner requests.
`pnpm test:integration` repeats those probes against all three running owner
compositions while validating all 31 business operations through the API and
RabbitMQ. Both hand-off gates are required; executed evidence belongs in
[verification](../verification.md).
