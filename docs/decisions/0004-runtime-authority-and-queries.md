# 0004 — Runtime authority, independent schemas and complete queries

Status: accepted, 29 September 2026. Refines the technical composition in 0002.
The aggregate, event-delivery and development-only rules remain in force.

## Workload authority

Each context retains a separate PostgreSQL database and credential. Two contexts
share each language service's process, which necessarily receives both sets of
credentials. Database grants enforce the authority of each login; they do not
isolate hostile code executing inside that same process. A stronger boundary
requires separate processes and secret delivery.

Only the topology bootstrap receives RabbitMQ administration credentials.
Runtime logins can publish their own event topics, consume their own queues and
use their own retry/dead-letter delivery exchange. They cannot declare or delete
topology. Delivery still uses confirms and commit-before-acknowledgement.

A small internal Go gateway holds Centrifugo's administrative API key. Each
context receives a distinct publisher credential permitting only its own
`cafe:<context>` channel. The gateway validates projection envelope identity and
forwards the original bytes. The session worker has a separate credential that
permits only disconnection. It cannot publish, and publishers cannot disconnect
users or access history/administrative methods. The gateway is technical
infrastructure and contains no domain policy.

Session Valkey retains durable revocation work with `appendfsync always`.
Realtime history runs in a separate disposable Valkey instance, with its own
credentials and ACL. History loss invokes reconciliation. Neither instance
evicts keys under memory pressure; resource sizing remains an operational concern.

Next.js builds a static export served by the existing ingress. Browser business
requests still use the Go API. There is no frontend Node process or frontend
secret authority.

## Secret delivery

Application composition accepts either `NAME` or `NAME_FILE`, with ambiguity and
empty values rejected. Development Compose mounts secrets as protected files
owned by the receiving runtime user. The launcher atomically writes generated
local configuration with mode `0600` and preserves existing credentials.

Paid development integrations use individual provider credentials, preferably
with provider-enforced spending controls. A non-production credential manager
may supply a protected file; configuration retains its path, not its contents.
The Docker daemon and launcher remain trusted. This interface does not implement
OpenBao deployment, leased credential renewal or production secret operations.
See [development secrets](../development-secrets.md).

## Query capabilities

Unpaginated list queries return complete arrays. Pagination is an explicit,
separate query capability with reusable application types and transport parsing
in Go, Python and TypeScript. `limit=1..100` and an optional resource-bound cursor
return `{items,nextCursor}`. Missing pagination parameters retain the original
query shape; invalid parameters are rejected rather than silently adjusted.

Keyset traversal uses stable UUID root identities. Cursors are versioned and
bound to the resource; they are not authorisation credentials. Browser helpers
support both ordinary queries and full page traversal. Initial and history-gap
reconciliation starts only after every server-authorised subscription attaches.
Live publications cover concurrent changes and per-root revision guards prevent
an older page from overwriting a newer publication. Pages do not constitute a
single database snapshot. Recoverable history needs no business query polling.

## Context persistence evolution

The two historical shared migrations remain frozen bootstrap infrastructure.
Each context now owns an additional manifest, ordered SQL migrations and its
own version sequence inside its language service. Initial local migrations add
indexes for the context's actual lifecycle query fields. Generic JSONB roots and
the existing atomic transaction contract remain the persistence model.

An administrative migration runner verifies ownership and immutable checksums,
serialises changes and commits each context's pending local migrations together.
Runtime users can read the identity and version ledger but cannot mutate it or
migrate their database. Generated service-local checksum metadata allows startup
to reject the wrong owner, changed checksums and missing or unknown versions.
One context can advance without advancing another context's migration sequence.

## Workflow visibility

Workers record structured, redacted failures and process-lifetime counters.
Authenticated diagnostics expose durable pending work separately from those
counters. Dead-letter transfer counts do not claim to be durable queue depth.
Liveness remains separate from workflow progress; a temporary broker outage does
not prevent commands from recording durable outgoing intent.

## Evidence

Required gates cover denied broker topology operations, cross-context realtime
publication denial, secret-file ambiguity and permissions, pagination beyond
100 roots, independent windows and history-gap recovery, and context migration
upgrade/rollback/identity/checksum checks. See [testing](../../TESTING.md) and
[executed verification](../verification.md) for the actual gate results.
