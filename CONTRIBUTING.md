# Contributing

Use British English. Start with [AGENTS.md](AGENTS.md), the
[architecture](ARCHITECTURE.md), [domain rules](DDD.md),
[browser contract](CLIENT-SUBSCRIPTIONS.md) and [testing guide](TESTING.md).
The [decision index](docs/decisions/README.md) maps the accepted rules and their
refinements. The [developer workflow](docs/developer-workflow.md) gives commands,
source examples and extension recipes.

## Plan the change around its owner

1. Run `pnpm context <name>` and open that context's README.
2. Name the business behaviour, immediate invariant and owning aggregate. Identify
   any subsequent workflow step that may remain pending.
3. Choose the application role: command, query, event reaction or projection.
4. Check which published contracts, private queued formats and persistence data
   the change affects. Preserve their identities and compatibility.
5. Add or revise the executable example and select the focused tests before wiring
   the change into the service.

Keep each context inside its owning service. Domain/application code depends on
plain types and application ports. Contexts communicate through published facts
or owner interfaces; each has its own database and credentials. The Go API and
Next.js translate and present business operations; domain policy stays with its owner.

## Put each responsibility where developers expect it

Paths in this table are relative to the owning context.

| Responsibility | Location and rule |
| --- | --- |
| Aggregate and its invariants | `domain/<business-name>`; children have no independent repositories |
| Command | `application/commands/<action>`; one named DTO and matching `CommandHandler` together |
| Query | `application/queries/<question>`; one named DTO and matching `QueryHandler` together |
| Read repository | `application/ports/<resource>_read_repository` (Go/Python) or `<resource>-read-repository` (TypeScript); name the interface/protocol `<Resource>ReadRepository` |
| Read representation | `application/readmodels/` (Go), `read_models/` (Python), `read-models/` (TypeScript) |
| Event reaction | One file in `application/event_handlers/` (Python) or `event-handlers/` (TypeScript) |
| Projection update | Go's current reactions live in `application/projections/` |
| Protobuf input mapping | `adaptors/messaging/`; HTTP mapping lives in `services/api/adaptors/http/backend/` relative to the repository root |
| Stored-state restoration and view mapping | `adaptors/postgres/` (Go) or `adaptors/persistence/` (Python/TypeScript) |
| DI bindings and worker/resource lifetime | Owning service's `apps/` composition module |

Use snake_case filenames in Go/Python and kebab-case in TypeScript. Import the
specific use-case package/module directly. Python package initialisers contain
documentation. Shared errors and event DTOs can have separate files; a new use
case starts as one file until it needs several supporting modules.

Name handlers by their role: `CommandHandler`, `QueryHandler`,
`IntegrationEventHandler`, `DomainEventHandler` or `ProjectionHandler`.
Aggregates retain business names. Document class/struct responsibilities and
important invariants with Go comments, Python docstrings or JSDoc. Concrete
adaptor names, or their Go package names, identify the implementation technology.
Use `<Resource>ReadRepository` for query persistence ports. Concrete classes use
`Postgres<Resource>ReadRepository`; Go factories use `postgres.New<Resource>ReadRepository`.
Use `<Aggregate>WriteRepository` for authoritative aggregate loads and saves.
Keep snapshot restoration and domain-fact publication mapping in owner adaptors.
The central command executor owns transactions, receipts and outgoing intent. See [decision 0015](docs/decisions/0015-explicit-persistence-role-names.md).
Keep handwritten source and documentation below 450 lines; record and explicitly
classify any justified exception in [the responsibility review](docs/large-file-review.md).

## Preserve execution and failure boundaries

A runtime aggregate business change starts in its named command handler. Make
the handler a plain repository consumer: load the target, invoke aggregate
behaviour and save. Register it through the central command executor in composition.
Feature handlers receive `CommandContext` and a business result; infrastructure
owns delivery metadata, deduplication, versions and transaction coordination. Commit one root, receipts, outcome and
outgoing event/realtime intent atomically. API commands also commit exact reply
bytes. Domain tests can exercise aggregate behaviour directly; rehydration,
projection updates and migrations have separate responsibilities.

Application handlers never invoke another command handler. An event reaction
maps a fact to an owner command and calls `DurableCommandPort`. Its receiving
transaction records acceptance, exact private Protobuf command bytes and dispatch
intent before ACK. The command queue then owns execution, bounded retries and
replay. This applies between roots in one context as well as across contexts.
Go's current subscriptions update projections; a new aggregate-changing Go
reaction requires the explicit durable-command port/adaptor and paired consumers.

Queries use application-owned read models and named read repository ports. Persistence
adaptors restore and validate stored authority before mapping it into a view.
Preserve absent-resource behaviour, revisions and continuation identities. An
unpaginated list remains complete; pagination is explicit.

Keep typed expected business rejections separate from retryable infrastructure
and corrupt-state failures. An identical command retry recovers its recorded
outcome. A new business attempt needs a new command ID. Preserve original receipt
material, expected versions and stable subscription/business identities;
correlation IDs only group workflow evidence.

## Compose plain handlers through DI

Use **Fx** in Go, **Dependency Injector** in Python and **Awilix** in TypeScript
services. Keep framework imports, registrations and container resolution under
`apps/`. Inject ordinary application ports into handlers; adaptors and composition
own PostgreSQL/RabbitMQ details. Resolve and test the full graph before workers
start, drain workers before closing pools, and test partial-startup cleanup.

Use the shared subscription helper to bind an event reaction and its private
command consumer. Keep typed subscription identifiers, codecs and the owner's
`adaptors/messaging/subscriptions.json` aligned. Validate the complete manifest
so missing, duplicate and mismatched registrations fail startup. Stable queue and
receipt identities survive class or file renames.

## Evolve sources, then generate and check

| Change | Source of truth | Required follow-through |
| --- | --- | --- |
| HTTP operation | `contracts/<context>/http_api/` | Update API HTTP translation, owner messaging mappings and frontend calls; run `pnpm generate:http` |
| Published command/query/event | `contracts/<context>/messaging/{commands,queries,integration_events}/` and owner envelopes in `messaging/v1/` | Update boundary decoders/mappings and run `pnpm generate:contracts` |
| Browser projection | `contracts/<context>/realtime/` | Update atomic publication and browser mapping; run `pnpm generate:contracts` |
| Private queued command/event | Owning context's `adaptors/messaging/` | Preserve stored-byte compatibility, codec tests and historical fixtures; run `pnpm generate:contracts` |
| Database evolution | Owning persistence adaptor's migration manifest and SQL | Append a checksummed migration and run `pnpm generate:contracts`; use administrative migration tooling |

Service/technical OpenAPI sources live under `contracts/services/<service>/http_api/`;
shared HTTP components live under `contracts/shared/http_api/`. Private domain
facts have no published contract version. RabbitMQ delivery alone does not make
them a public contract. Database bootstrap lives under `devops/postgres/bootstrap/`.

Commit the generated outputs and review their diffs. Edit their sources rather
than generated files. Required Protobuf scalars have explicit presence and the
`required_input` annotation; optional additions stay optional and zero remains a
valid supplied value. Generated types stay in adaptors; the API-to-owner transport
is RabbitMQ, with anti-corruption mapping at both ends. Frontend HTTP calls use
the generated operation client.

`pnpm check:contracts` rejects changed, missing and newly generated outputs.
`pnpm check:compatibility --against <commit>` separately compares historical
interfaces and private fixtures. Breaking public changes need a version transition.
See [the contract guide](contracts/README.md) for sources and generated locations.

Runtime credentials only verify database identity, schema versions and checksums.
They never run migrations. Preserve the frozen bootstrap SQL and existing migration
history. `pnpm dev:up` applies pending migrations through administrative tooling.

## Test, document and hand off

Put Go/TypeScript tests beside their subjects. Mirror Python context tests under
`tests/contexts/<context>/`; retain native Gherkin bindings under the service test
tree. Business-rule changes need named examples in `specifications/<context>/`
and their native bindings. Use deterministic values in unit tests; prove delivery,
receipts, concurrency and crash recovery with real infrastructure.

Run focused tests during editing, then both hand-off gates:

```sh
pnpm test:focused <service> --context <context>
pnpm verify
pnpm test:integration
```

`pnpm verify` covers all five applications, generation/compatibility, architectural
negative fixtures, DI graphs, type checks and fast scenarios. `pnpm test:integration`
provisions and removes a disposable PostgreSQL/RabbitMQ/Centrifugo/Valkey/browser
stack. Report actual results and any unavailable lane in
[executed verification](docs/verification.md). A fast pass does not establish
transaction or delivery correctness.

When an architectural rule changes, update its accepted decision, contributor
and agent guidance, executable examples and regression checks together. Update
context navigation when ownership or entry points change. Record notable changes
under `Unreleased` in [CHANGELOG.md](CHANGELOG.md).

Use conventional commits such as `feat(loyalty): add reward cancellation`.
Describe the triggering behaviour, final rule and evidence in the PR. Keep secrets,
`.local`, `.tools` and build outputs out of commits. This reference supports
local/development compositions only; release workflows are outside its scope.
