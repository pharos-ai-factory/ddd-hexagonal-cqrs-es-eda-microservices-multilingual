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
PostgreSQL server is an operational convenience. Contexts cannot read each
other's tables. A service is a composition root for two independent contexts;
co-location does not widen a command's authority.

The source layout keeps explicit dependency roles inside service boundaries:

```text
services/storefront/{apps,contexts,foundation,contracts,integrations}/  Go module
services/operations/src/operations/{apps,contexts,foundation,contracts,adaptors}/ Python
services/engagement/src/{apps,contexts,foundation,adaptors,contracts}/ TypeScript
services/api/{apps,application,adaptors}/                             Go module
services/web/src/{app,features,adaptors}/                             Next.js
contracts/{events,realtime,persistence,http}/  shared wire/schema specifications
specifications/{menu,ordering,...,workflows}/  executable Gherkin behaviour
tests/acceptance/                             cross-service wire/API bindings
devops/                                      development composition only
scripts/                                     gates and executable journeys
docs/                                        accepted decisions and exercises
```

Go and TypeScript unit tests sit beside their subjects. Python tests and
TypeScript Gherkin bindings live under their service's `tests` directory.
Go Gherkin bindings sit beside the owning application handlers.
Generated contracts stay in transport/adaptor rings.
The shared root has no cross-language business implementation. Application
handlers may use plain published DTOs, never generated transport messages.
The shared Gherkin catalogue has native service runners and a separate live
workflow runner. Test decision probes make no transaction or delivery claims;
real infrastructure proves those boundaries. See decision 0003 and `TESTING.md`.

Operations uses explicit `TypedDict` commands, published DTOs, outcomes and
snapshots, with generic ports bound to each aggregate's snapshot type. Domain
facts are immutable dataclasses owned by their context. HTTP, Protobuf and
PostgreSQL adaptors validate external values before exposing typed application
values. The consumer retains the original wire payload for receipt fingerprints.
Strict Mypy checks cover handwritten runtime code and native tests; generated
Protobuf type declarations stay in the adaptor ring.

The Go API owns technical sessions in Valkey. It holds no business database or
RabbitMQ credentials, and it cannot decide domain outcomes. It routes only known
context paths, authenticates the caller, replaces browser credentials with the
owning service's internal key, and preserves command identity and version headers.

Next.js renders the operator interface. An ingress routes `/auth` and `/api/v1`
to the Go API and `/connection/websocket` to Centrifugo. Internal proxy and
administrative endpoints are not exposed through browser ingress.

## Transaction rule

One command or event-handling transaction changes at most one aggregate,
together with its receipt, recorded outcome and outgoing events. Application
ports expose one aggregate operation and never expose a SQL transaction or an
arbitrary collection of repositories.

If an invariant requires two pieces of business state to change immediately,
they belong inside one consistency boundary. Extract unrelated lifecycles
before enlarging an aggregate.

Workflows advance through separate transactions with explicit pending states.
Neither an in-memory callback nor a synchronous call to a second handler is a
durable workflow hand-off.

## One delivery mechanism

Delivered domain and integration events follow the same pipeline:

```text
aggregate + receipt + outcome + immutable event outbox
                    → PostgreSQL commit
leased dispatch → persistent mandatory RabbitMQ publication → publisher confirm
consumer → one aggregate + receipt + outgoing events → PostgreSQL commit → ACK
```

Domain facts are plain types. Applications map them to private domain contracts
or published integration contracts. Both have versioned Protobuf representations,
stable message identities, correlation and causation. Context-private messages
are only bound to consumers owned by that context.

Immutable source events and mutable dispatch progress are separate. A relay
lease is fenced: an expired publisher cannot complete a subsequently reclaimed
dispatch. Consumers tolerate duplicate publication and redelivery. Poison or
exhausted messages enter a named dead-letter queue and require an explicit replay.

The broker does not participate in either database transaction. A lost commit
response is an unknown outcome resolved by idempotent retry. A queue being empty
does not establish business completion.

## Reads and external effects

Query handlers return DTOs without changing aggregates. Consumer-owned projections
contain only published facts. Published menus remain immutable and orderable in
this example; introducing withdrawal would require an explicit owner protocol.

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
Go, Python and TypeScript independently verify database identity, restricted
privileges and migration checksums. No runtime credential can migrate a schema.

Only local/development configurations are supported. Runtime processes reject
staging and production environment values. Schema administration is a separate
bootstrap responsibility; application startup cannot migrate a database.
