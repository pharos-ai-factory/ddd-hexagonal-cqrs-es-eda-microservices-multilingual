# Café reference architecture

This repository implements an independently runnable, development-only
click-and-collect café. The service-first monorepo contains Go Storefront, Python Operations and
TypeScript Engagement, plus a separate Go API and a Next.js frontend. Versioned
wire contracts and behavioural scenarios connect their independent builds.

## Ownership and deployment

| Service | Contexts | Authority |
| --- | --- | --- |
| Storefront (Go) | Menu, Ordering | Published drink/menu revisions and accepted orders |
| Operations (Python) | Preparation, Collection | Production work and physical handover |
| Engagement (TypeScript) | Loyalty, Customer Communication | Earned grants, usable rewards and notification delivery |

Each context owns a PostgreSQL database and a distinct runtime login. A shared
PostgreSQL server is an operational convenience. Each credential is denied access
to the other context databases. A service is a composition root for two contexts
and holds both credentials: the process is their shared security boundary.
Co-location does not widen a command's authority. Isolation from hostile code in
a sibling context would require distinct processes and secret delivery.

The source layout keeps explicit dependency roles inside service boundaries:

```text
services/storefront/{apps,contexts,foundation,contracts,integrations}/  Go module
services/operations/src/operations/{apps,contexts,foundation,contracts,adaptors}/ Python
services/engagement/src/{apps,contexts,foundation,adaptors,contracts}/ TypeScript
services/api/{apps,application,adaptors}/                             Go module
services/web/src/{app,features,shared,adaptors}/                      Next.js
contracts/<context>/{messaging,realtime,http_api}/  published specifications
specifications/{menu,ordering,...,workflows}/  executable Gherkin behaviour
tests/acceptance/                             cross-service wire/API bindings
devops/                                      development composition only
scripts/                                     gates and executable journeys
docs/                                        accepted decisions and exercises
```

Go and TypeScript unit tests sit beside their subjects. Python context tests mirror
the source under `tests/contexts/<context>/`; Python and TypeScript Gherkin bindings
live under their service's `tests` directory.
Go Gherkin bindings sit beside the owning application handlers.
Generated contracts stay in transport/adaptor rings.
Each context has `application/commands/` and `application/queries/`, with one
named DTO and its matching handler per business use-case file. Go/Python filenames
use snake_case; TypeScript uses kebab-case. Callers import those packages/modules
directly. Event reactions and projection handlers have separate business-named
files. Read models and named read repository ports belong to the application; decoding,
restoration and view mapping belong to context adaptors. Shared infrastructure
contains technical behaviour. See decisions
[0013](docs/decisions/0013-command-module-layout.md) and
[0014](docs/decisions/0014-context-navigation-and-query-layout.md).
Published Protobuf payload sources identify their context and interface category.
Context folders contain `messaging/{commands,queries,integration_events}`,
`realtime` and `http_api`. Private domain facts are plain application/domain types;
private RabbitMQ formats and delivery metadata remain in owner service adaptors.
The public integration envelope excludes private payloads. Private work retains
stored-byte compatibility without becoming a published interface. Commands and
queries use Protobuf over RabbitMQ between the API and owner services; OpenAPI
defines the public HTTP boundary. See decision 0008 and `contracts/README.md`.
The shared root has no cross-language business implementation. Application
handlers may use plain published DTOs, never generated transport messages.
The shared Gherkin catalogue has native service runners and a separate live
workflow runner. Test decision probes make no transaction or delivery claims;
real infrastructure proves those boundaries. See decision 0003 and `TESTING.md`.

Operations uses explicit `TypedDict` commands, published DTOs, outcomes, snapshots
and read views. Command ports retain the owning snapshot type; named read repositories
expose application-owned views. Domain
facts are immutable dataclasses owned by their context. HTTP, Protobuf and
PostgreSQL adaptors validate external values before exposing typed application
values. The consumer retains the original wire payload for receipt fingerprints.
Strict Mypy checks cover handwritten runtime code and native tests; generated
Protobuf type declarations stay in the adaptor ring.

The Go API owns technical sessions in Valkey and has no business database. It
maps only known OpenAPI operations, authenticates the caller and translates HTTP
requests into typed Protobuf commands/queries over RabbitMQ. Owner-side adaptors
translate those messages into plain application types. A dedicated API broker
credential can publish requests and consume replies; owner credentials retain
context authority. Command identity, expected version and correlation survive
translation. Success reports a committed owner outcome; a timeout remains uncertain
and permits an identical retry. Owner request packages, explicit ACL mappings,
separate command/query consumers and transactionally stored response bytes are
defined by decision 0009. See decision 0007 and `contracts/shared/messaging/requests.md`.

Owner services expose only health and authenticated diagnostics over HTTP.
Their composition injects typed executors and query handlers directly into the
RabbitMQ request registry. Contexts have messaging adaptors for incoming business
requests. The earlier direct owner HTTP inspection endpoints were retired by
[decision 0016](docs/decisions/0016-owner-messaging-boundary.md); local inspection
uses the public API and the same broker path.

The API's HTTP adaptor has separate `operational`, `session`, `realtime` and `backend`
packages with tests beside their handlers. The parent package composes those
routes and verifies the assembled OpenAPI surface. Shared response and security
helpers stay under `http/internal`; each route package receives its own technical
dependencies without importing the parent composition.

OpenAPI business fragments live in `contracts/<context>/http_api/`. Complete
service documents and technical routes live in `contracts/services/<service>/http_api/`.
Reusable components live in `contracts/shared/http_api/`. Go modules embed
self-contained generated bundles, require complete route registration at startup
and validate deterministic/live response conformance. See decision 0006 and
`contracts/shared/http_api/README.md`.

Next.js exports the operator interface as static assets served by the ingress;
there is no frontend Node server. The ingress routes `/auth` and `/api/v1`
to the Go API and `/connection/websocket` to Centrifugo. Internal proxy and
administrative endpoints are not exposed through browser ingress.

An internal Go realtime gateway authorises each context credential for publication
to its own channel and the session credential for disconnection. Domain processes
do not receive Centrifugo's full API key. RabbitMQ runtime credentials cannot
create, change or delete topology; the separate bootstrap owns those operations.
Session/revocation state and disposable realtime history use separate Valkey
instances and credentials. Session writes retain `appendfsync always`; history
loss is handled by browser reconciliation.

Runtime credentials arrive through explicit secret-file references. The local
launcher also supports protected files containing individual developer credentials
without copying their contents into configuration. See `docs/development-secrets.md`.

Each service exposes authenticated `/diagnostics` on its internal/loopback HTTP
endpoint, using its existing bearer credential; the API uses its CLI key.
Database metrics include pending outbox/realtime publications, oldest age and
failed pending dispatches. Session metrics include pending revocations and expired
sessions awaiting revocation. Worker failure records omit exception messages,
credentials and payloads. Their counters reset on process restart; confirmed
dead-letter transfers are not queue-depth measurements. These routes are absent
from browser ingress. `/healthz` reports liveness separately from workflow progress.

## Transaction rule

Every runtime aggregate business transition starts in a named owner command
handler. Its `Execute`/`execute` method loads the target through its write
repository, changes that root and saves it. The central command executor wraps
the invocation and commits receipts, the outcome and outgoing event/realtime intent.
API command processing includes exact reply bytes. Application ports expose one
aggregate operation and never expose a SQL transaction or an arbitrary collection
of repositories.

Application handlers never invoke another command handler. An event reaction
persists an owner command through `DurableCommandPort`; a separate command consumer
executes it. Projection updates, rehydration, domain tests and migrations retain
their distinct roles. Decisions 0011 and 0012 define this stricter convention.

If an invariant requires two pieces of business state to change immediately,
they belong inside one consistency boundary. Extract unrelated lifecycles
before enlarging an aggregate.

Workflows advance through separate transactions with explicit pending states.
Neither an in-memory callback nor a synchronous call to a second handler is a
durable workflow hand-off.

## One delivery mechanism

Delivered domain and integration events share the delivery infrastructure.
An aggregate-changing reaction follows two receiving stages:

```text
producer command → aggregate + receipts + outcome + event/realtime intent → commit
event outbox → persistent mandatory RabbitMQ publication → publisher confirm
event reaction → acceptance + exact owner command bytes + dispatch → commit → event ACK
command outbox → persistent mandatory RabbitMQ publication → publisher confirm
command consumer → one aggregate + receipts + outcome + publications → commit → command ACK
```

The receiving outbox makes acceptance durable before command execution. Failure
before acceptance commits remains the event consumer's retry responsibility. After
acceptance, the command's dispatch and consumer own recovery. Command execution
failure cannot erase the already persisted intent. Projection consumers commit
their projection and receipt before ACK without introducing an aggregate command.

Domain facts are plain types. Applications map them to owner-private delivery values or published integration
events. Public interfaces have versioned Protobuf representations. Internal
formats preserve stored delivery bytes and keep their definitions with the owner.
Both deliveries retain stable identities, correlation and causation. Context-private messages
are only bound to consumers owned by that context.

Immutable source events and mutable dispatch progress are separate. A relay
lease is fenced: an expired publisher cannot complete a subsequently reclaimed
dispatch. Consumers tolerate duplicate publication and redelivery. Poison or
exhausted messages enter a named dead-letter queue and require an explicit replay.

The broker does not participate in either database transaction. A lost commit
response is an unknown outcome resolved by idempotent retry. A queue being empty
does not establish business completion.

## Feature command execution

Command handlers depend on named write repositories and a target-only
`CommandContext`. They load an aggregate, invoke its behaviour and save it.
The aggregate records private domain events; owner adaptor mappers stage the
selected outgoing messages when the repository saves.

A central executor supplies a fresh repository to each invocation and owns the
local PostgreSQL transaction, receipt checks, expected versions, outcomes and
outgoing bytes. Rejected commands can commit a durable rejection without saving
an aggregate. Infrastructure failures roll back the whole attempt. There is no
application `UnitOfWork` dependency or transaction callback. Each transaction is
restricted to one aggregate in one context database; cross-context work proceeds
through RabbitMQ. Composition binds the shared executor once per use case.

## Reads and external effects

Each query has a named input, handler, read repository port and application-owned read model.
Owner RabbitMQ query adaptors construct that input after the API translates HTTP to Protobuf. Persistence read repositories validate
stored aggregate authority and select its read fields; query handlers and read
models stay independent of aggregate implementations. The current read repositories use
stored snapshots; specialised read tables can implement the same ports later.
Ports use `<Resource>ReadRepository`; Python/TypeScript implementations include
`Postgres`, while Go uses the package-qualified `postgres.New<Resource>ReadRepository`
factory. Snapshot restoration is independently reusable by read repositories and
write repositories. See [decision 0015](docs/decisions/0015-explicit-persistence-role-names.md).
Consumer-owned projections contain only published facts. Published menus remain immutable and orderable in
this example; introducing withdrawal would require an explicit owner protocol.

List queries can return complete arrays or explicitly opt into keyset pagination.
The browser follows every page after all subscriptions attach and after a history
gap. No query silently truncates its results. Pagination is a separate query-port
capability and transport tooling rather than domain policy.

Notification delivery uses an idempotent development provider. Provider I/O is
outside aggregate transactions; recording its outcome is a separate command.
The provider's idempotency receipt protects the crash window after acceptance.

## Scope

Current-state persistence is authoritative. Event history provides publication
and audit evidence; aggregates are not reconstructed from that history.

The Next.js frontend uses Centrifugo with Protobuf browser projections. Context
transactions also write exact browser projection bytes and a realtime dispatch
intent atomically. A fenced relay publishes those bytes to Centrifugo. Server-side
subscriptions come from the Go API's session authority; each window reconciles
from owner queries at initial attachment or a history gap. Recoverable history
requires no additional business GET. See `CLIENT-SUBSCRIPTIONS.md`.

Database administration applies checksummed migrations before runtime startup.
The two historical SQL bootstrap files in `devops/postgres/bootstrap/` are frozen. Each context owns its additional
manifest, SQL and independent version sequence inside its language service.
The administrator records a context identity and immutable checksum ledger;
initial local migrations add context-specific lifecycle indexes.
Go, Python and TypeScript independently verify database identity, restricted
privileges and migration checksums. No runtime credential can migrate a schema.
Append a context-local SQL migration and manifest checksum, then run
`pnpm generate:contracts`. `pnpm dev:up` applies pending migrations; the explicit
administrative command is `python3 scripts/migrate.py --env-file .local/dev.env`,
optionally with `--owner ordering` to advance one context only.

Only local/development configurations are supported. Runtime processes reject
staging and production environment values. Schema administration is a separate
bootstrap responsibility; application startup cannot migrate a database.


Frontend HTTP calls use OpenAPI-generated operation/request/response types through
the HTTP adaptor. Verification rejects generated drift and incompatible feature
usage. Command replies use owner-local immutable storage and fenced confirmed
publishers; their thirty-second lifetime bounds retries. Exact reply bytes commit
atomically with the receiving outcome. See decision 0009 and the contract guides.

## Explicit composition and durable commands

[Decision 0010](docs/decisions/0010-dependency-injection-frameworks.md) records the
framework comparison and selection: Go Fx, Python Dependency Injector and
TypeScript Awilix. Frameworks and concrete resource bindings live in `apps/`.
Application handlers receive plain ports. Context containers own separate pools
and credentials. Provider-graph tests resolve every binding; workers stop and
drain before resource disposal, including after partial startup failure.

[Decision 0011](docs/decisions/0011-durable-commands-before-aggregate-mutations.md)
requires named commands for aggregate business transitions. Event handlers map
facts to durable owner commands. Their receiving transaction records exact private
Protobuf command bytes and a dispatch intent; RabbitMQ command consumers own
execution, retries and replay. Projection handlers retain their read-model role.
`CommandHandler`, `IntegrationEventHandler`, `DomainEventHandler` and
`ProjectionHandler` suffixes identify application responsibilities. Domain roots
retain their business names. Transport and persistence classes identify their
concrete adaptor role; Go package-qualified types supply that infrastructure scope.

[Decision 0012](docs/decisions/0012-enforcement-and-developer-workflow.md) strengthens
command entry-point enforcement, partial-startup cleanup and exhaustive typed
subscription registration. `pnpm dev:status` checks authority connections and
actual broker consumers; `dev:up` waits on it. Historical compatibility is checked
against an accepted revision independently from generated-file drift.

## Finding a business capability

Each context has a source-root README. `pnpm context <name>` lists its current
commands, queries and reactions, plus contracts, specifications and its focused
test command. `docs/contexts.json` records the six owners and their browser features.
Application queries have named inputs, handlers, read models and read repository ports;
context adaptors restore stored authority and map transport arguments. Individual
event reactions and domain aggregates have business-named files. Browser features
follow user tasks, with reusable delivery/recovery code in `shared/`.
See [decision 0014](docs/decisions/0014-context-navigation-and-query-layout.md).
