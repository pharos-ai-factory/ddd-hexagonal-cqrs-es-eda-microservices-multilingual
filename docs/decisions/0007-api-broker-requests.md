# 0007 — API command and query dispatch through RabbitMQ

Status: accepted, 8 October 2026. Supersedes the API HTTP forwarding and absence
of broker authority in 0002. Refines request authority in 0004 and the mapping of
OpenAPI operations in 0006. Aggregate, event-delivery and browser rules continue. Decision 0009 subsequently
refines owner request packages, queue separation and durable reply acknowledgement.

## Decision

Keep HTTP JSON and OpenAPI as the browser/CLI boundary. The Go API authenticates
requests, validates their structure and translates them into explicit versioned
Protobuf commands or queries. Internal API-to-context business communication uses
RabbitMQ. Owner-side adaptors validate those messages and construct plain
application types. Neither boundary places business decisions in the API or
imports generated messages into application/domain packages.

Place command, query and query-reply sources under the owning context in
`contracts/<context>/messaging/{commands,queries}/v1/`. Keep shared envelopes and
technical metadata under `contracts/shared/messaging/v1/` (0008). Generated Go bindings are built
independently into both Go modules; Python bindings and TypeScript descriptors
remain in the consuming service's adaptor ring. Exact generated paths are derived
by the architecture checker and have the same compiler-owned size exception as
other Protobuf bindings. Query replies have separate DTOs from browser snapshots.

The bootstrap declares durable quorum request/reply queues. A dedicated API
credential publishes requests and consumes replies; each context can consume its
request queue and publish its reply topic. No runtime credential can change
topology. Owner reply queues identify the authority of each reply. The development
composition has one API process; exclusive reply consumers enforce that scope.
Concurrent requests within that process use separate transport identities.

## Outcomes and recovery

The HTTP call waits for an owner reply. Command success reports the committed
owner outcome. Queries return the owner's read DTOs through the same broker path.
The two request types retain separate handlers and contracts. Using request/reply
for the external interaction does not synchronously invoke another aggregate's
handler from an application command handler; downstream workflow reactions still
advance through durable domain/integration events.

Persistent mandatory publications require confirms. Owners acknowledge after
handling and confirmed reply publication. A crash between commit and acknowledgement
can redeliver a command; its existing command receipt recovers the recorded
outcome and protects outgoing intent from duplication. Invalid messages are
quarantined. Infrastructure failures return a retryable technical result;
broker redelivery after connection failure has a bounded quorum delivery limit.

Command IDs and expected versions survive boundary translation. A fresh transport
request ID matches each reply and never enters receipt deduplication. Broker
acceptance alone cannot establish command completion. A timeout returns 503,
leaves completion uncertain and retains the browser's original attempt for an
explicit identical retry. Waiting requests and stale replies have bounded TTLs;
expiry cannot cancel a transaction already executing.

HTTP schemas remain authoritative for public request/response shapes. Startup
registration and conformance tests still cover complete OpenAPI operation mapping.
Direct owner HTTP endpoints remain development inspection interfaces. Technical
sessions, Centrifugo callbacks/publication and notification-provider I/O keep their
existing dedicated transports.

## Evidence

Deterministic tests cover all 31 API operations, required field presence, zero
prices/versions, plain application types, resource-bound pagination and malformed
or mismatched replies. Live workflows exercise each language's owner consumers.
Broker permission probes deny API event publication, owner command publication,
foreign replies and foreign request consumption. A real reply-loss fixture proves
a command can commit while HTTP reports uncertainty, then return the same outcome
on retry with one receipt, one event and one root transition.

Executed results and environment limits are recorded in [verification](../verification.md).
