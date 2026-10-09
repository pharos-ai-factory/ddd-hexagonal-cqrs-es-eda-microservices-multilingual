# Reference implementation rules

Use British English and concise, direct prose. Read [ARCHITECTURE.md](ARCHITECTURE.md),
[DDD.md](DDD.md), [CLIENT-SUBSCRIPTIONS.md](CLIENT-SUBSCRIPTIONS.md) and
[TESTING.md](TESTING.md) before architectural changes. Follow
[CONTRIBUTING.md](CONTRIBUTING.md) and the [developer workflow](docs/developer-workflow.md).
The [decision index](docs/decisions/README.md) explains which accepted decisions
refine earlier rules. Decisions 0010–0015 define DI, durable command reactions,
enforcement, the source layout and persistence role names.

## Start with ownership

- Run `pnpm context <name>` and read the context README before editing a use case.
  Name its invariant, aggregate, command/query and allowed pending workflow state.
- Keep contexts inside their owning language service. Do not import another
  service's source or place business policy in the Go API or Next.js.
- Domain and application packages are independent of transport, persistence and DI.
  Domain packages never import another context. Context application packages
  never import another context's implementation.
- Contexts own separate PostgreSQL databases and credentials. Inter-context
  reads use consumer-owned projections or published owner ports.
- Aggregate children have no independent repositories. Reconsider the model
  when immediate business invariants span proposed aggregates.

## Commands, queries and reactions

- Every runtime aggregate business transition, including creation, starts in a
  named owner command handler. Inject its named write repository, load the target,
  invoke aggregate behaviour and save. The central command executor owns the
  invocation transaction; feature code has no transaction callbacks.
  Domain tests, read-only rehydration, projections and migrations have distinct roles.
- Application handlers never invoke another command handler. Integration and
  private domain event handlers enqueue owner commands through `DurableCommandPort`.
  Persist receiving evidence, exact private command bytes and dispatch intent
  before ACK. The command consumer owns subsequent execution and retries.
- Delivered domain events and integration events use the same transactional
  outbox, RabbitMQ publisher confirms, manual acknowledgements, receipts,
  bounded retries, dead-letter and replay mechanisms.
- Commit one aggregate, receipts, outcome and outgoing event/realtime intent
  atomically. API commands also persist exact reply bytes. ACK after commit.
- Query handlers use named application read repository ports and application-owned read
  models. Context persistence adaptors validate stored authority and map views.
  Preserve missing-resource behaviour, revisions and complete/paginated reads.
- Use stable identifiers, immutable published revisions, typed business rejections
  and expected versions for caller-driven commands. Identical retries retain their
  command identity and input. Infrastructure/corrupt-state failures stay retryable.
  Correlation IDs group workflows and never provide deduplication.

## Source layout and composition

- Put one command DTO/handler pair in `application/commands/<action>` and one
  query DTO/handler pair in `application/queries/<question>`. Import their concrete
  modules directly. Use snake_case files in Go/Python and kebab-case in TypeScript.
- Give each event reaction or projection handler its own business-named file.
  Python uses `event_handlers` and `read_models`; TypeScript uses `event-handlers`
  and `read-models`; Go uses `projections` and `readmodels` for its current roles.
  Keep Python package initialisers documentation-only.
- Keep decoding, restoration and business mappings in the owning context's
  adaptors. Shared infrastructure stays technical; generated catalogues may be shared.
- Name query persistence ports `<Resource>ReadRepository`; use
  `Postgres<Resource>ReadRepository` in Python/TypeScript and
  `postgres.New<Resource>ReadRepository` constructors in Go. Keep snapshot
  restoration separate. Use `<Aggregate>WriteRepository` in command handlers;
  central command executors own transactions and delivery bookkeeping.
- Use explicit role names (`CommandHandler`, `QueryHandler`, `IntegrationEventHandler`,
  `DomainEventHandler`, `ProjectionHandler`) and document responsibilities with
  Go comments, Python docstrings or JSDoc. Aggregates retain business names.
- Use Fx in Go, Dependency Injector in Python and Awilix in TypeScript services.
  Framework imports, container resolution and registrations belong under `apps/`.
  Handlers receive plain constructor ports. Resolve graphs before workers start;
  drain workers before disposal and clean up resources after partial startup failure.
- Register each durable event/command consumer pair through the shared runtime
  helper. Keep stable typed subscription IDs aligned with owner `subscriptions.json`
  and validate the complete set. Go's current reactions update projections; a new
  aggregate-changing Go reaction needs the explicit durable-command implementation.

## Published boundaries and browser behaviour

- The API translates OpenAPI HTTP JSON into Protobuf commands/queries over
  RabbitMQ. Owner adaptors translate wire messages into plain application DTOs.
  Owner services expose only health and authenticated diagnostics over HTTP.
  Register business handlers with messaging adaptors; keep HTTP routes in the API.
- Published sources use `contracts/<context>/{messaging,realtime,http_api}`.
  Messaging groups published commands, queries and integration events.
- Keep domain facts separate from versioned transport envelopes. Domain events
  remain private to their context even when they travel through RabbitMQ.
  Private queued formats and compatibility fixtures live in owner adaptors.
- Generate with `pnpm generate:contracts` after Protobuf/private-message changes,
  or `pnpm generate:http` for OpenAPI alone. Commit generated outputs and preserve
  historical wire identities/fixtures. Compare compatibility as well as generated drift.
- Browser updates use server-authorised Centrifugo subscriptions and the separate
  Protobuf projection contract. No browser polling or client-selected channels.
- Persist exact realtime bytes atomically with the root transition. Test recovery,
  stale revisions, independent windows and durable session revocation.
- Browser features follow business tasks; reusable command recovery/realtime code
  stays under `shared/`. HTTP calls use the generated operation client. Preserve
  uncertain command attempts across reload and authentication loss.

## Verify and hand off

- Keep current-state persistence. A change to event sourcing requires an explicit
  decision. Runtime credentials never run migrations; bootstrap remains administrative.
- All runtime compositions are development-only. No staging or production
  release workflow is permitted.
- Keep handwritten files below 450 lines. Review and record justified exceptions;
  generated files are explicitly classified.
- Put Go/TypeScript tests beside their subjects. Mirror Python context tests under
  `tests/contexts/<context>/`; retain native Gherkin bindings. Use deterministic
  identities/timestamps; real infrastructure proves transaction and delivery claims.
- Start with `pnpm test:focused <service> --context <context>`. Scaffolds under
  `.local/scaffolds` require reviewed policy, registration and behaviour tests.
- Update the affected guides, accepted decision, context navigation and executable
  examples together. Add negative fixtures when changing an enforced rule.
- Run `pnpm verify` and `pnpm test:integration` before hand-off. Report any
  skipped environment-dependent checks and record executed evidence in
  `docs/verification.md`. Keep secrets and build artefacts out of commits.
