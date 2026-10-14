# 0020 — Shared Operations adaptor responsibilities

Accepted: 14 October 2026. Refines decisions 0014–0017 and 0019.

## Decision

Group the shared Python Operations adaptors by technical responsibility:
`persistence/`, `messaging/`, `realtime/` and `http/`. Retain shared command execution,
workflow diagnostics and generated outputs at the adaptor root. Business use
cases and their mappings remain in the owning Preparation or Collection context.

Use specific module names: `event_codec`, `private_event_codec`,
`internal_command_codec`, `rabbitmq_request_registry`, `rabbitmq_consumer`,
`rabbitmq_outbox_relay`, `rabbitmq_reply_relay`, `projection_codec` and
`gateway_relay`. Persistence modules identify snapshots, aggregate transactions,
command/reply/integration outboxes and dispatch leases. Concrete repository and
transaction classes retain their PostgreSQL role names from decision 0015.

Split database persistence from broker and realtime delivery. Persistence owns
lease claims/completion and exact stored bytes. Relays own network I/O and call
those persistence capabilities. The generic command executor still coordinates
one named handler; the aggregate transaction still commits one root, receipts,
outcome, private facts, realtime intent and API reply atomically.

Generate the Python request/reply type catalogue and owner factories from the
service's declared request families. Shared envelope validation and request
consumption use that catalogue. Context messaging adaptors retain the mappings
between wire payloads and named application commands and query views.

Package initialisers contain documentation only. Callers import concrete modules.
The shared adaptor README explains the responsibility map and traces a completed
ticket through private delivery and independent integration publication. Context
navigation links to that map and keeps business entry points prominent.

## Behaviour and compatibility

Preparation's ticket mapper stages `preparation.drinks-ready.domain` from
`DrinksReadyDomainEvent`. Its independently delivered publication handler maps
that fact to `DrinksReadyIntegrationEvent`, following decision 0019. This corrects
a stale mapper which referenced the former domain/contract type names and called
the integration publication factory during aggregate saving.

The reorganisation retains published contracts, private queued formats,
subscription/receipt identities, database schemas, transaction boundaries,
publisher confirms, acknowledgement order, retry limits and lease fencing.
Generated output locations stay under `generated/`. Separate context databases
and credentials remain authoritative.

## Trade-offs and enforcement

The additional modules expose the existing persistence and delivery boundaries.
Readers can locate a codec, database operation or network worker from its name.
The service-level structure remains technical; the contexts express the business
capabilities. Shared diagnostics intentionally assemble backlog and worker evidence.

Architecture checks reject new flat shared adaptor modules outside the documented
root roles. Negative fixtures reject direct database drivers in shared messaging
or realtime adaptors and network clients in persistence. Shared event-codec checks
follow the relocated modules. Generated-drift checks cover the request catalogue.

Native tests exercise completed-ticket private encoding and owner request factories.
Existing PostgreSQL/RabbitMQ tests exercise transaction rollback, immutable bytes,
reply recovery, stale leases and durable delivery through the relocated modules.
Run focused Operations tests, `pnpm check --changed`, `pnpm verify` and
`pnpm test:integration`; record executed evidence in `docs/verification.md`.
