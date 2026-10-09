# Contributing

Use British English. Read `AGENTS.md`, `ARCHITECTURE.md`, `DDD.md` and `TESTING.md`
before changing the model or infrastructure.

## Change one responsibility

Identify the owning context, aggregate and command or query first. Explain the
immediate invariant. An aggregate child has no independent write repository.
A cross-aggregate reaction requires a recorded asynchronous hand-off, including
when both aggregates run in the same process.

Composition roots select implementations and bind handlers. Notification content,
event interpretation and projection policy belong in the owning application
package. Domain packages contain neither transport contracts nor provider calls.

Run:

```sh
pnpm verify
pnpm test:integration
```

When changing a business rule, add or revise a named example in
`specifications/<context>` and its native step binding. Start with
`pnpm test:bdd` for fast feedback; use the infrastructure lane for delivery,
receipt or concurrency claims. Follow the [Gherkin guide](specifications/README.md)
for scenario IDs and boundary selection.

The first command checks all five applications, including both Go modules,
Python and TypeScript. Python uses pinned Mypy in strict mode for all handwritten
source and tests, including Gherkin bindings and infrastructure fixtures.
Run `uv run --frozen mypy` inside `services/operations` for focused feedback.
Preserve typed command/event/snapshot boundaries; narrow external values through
runtime decoders instead of casting them into trusted application types.
The second command creates and removes an isolated Docker project
and includes PostgreSQL, RabbitMQ, Centrifugo, Valkey and Chromium evidence. Report any omitted lane explicitly.

Use conventional commits such as `feat(loyalty): add reward cancellation` or
`fix(amqp): preserve delivery identity during replay`. Pull requests should
describe the triggering behaviour, the final rule and the relevant evidence.
Record notable changes under `Unreleased` in [the changelog](CHANGELOG.md).
Do not include generated secrets, `.local`, `.tools` or executable build outputs.

## Contracts and persistence

Generated Go/Python Protobuf bindings, Python `.pyi` declarations and TypeScript
schema descriptors are committed. To regenerate them, install `uv` and
run `pnpm generate:contracts`; the compiler and Go plugin versions are pinned in
`scripts/generate.py`. Normal builds do not require the generator.

Review schema and fixture changes as public contract changes. See
`contracts/shared/messaging/events.md` for compatibility rules. Regeneration must produce
no unexpected change before merging. `pnpm check:contracts` regenerates in an
isolated directory and fails on changed, missing and newly generated outputs.

Find published payloads under `contracts/<context>/messaging/` or
`contracts/<context>/realtime/`. Keep public package names and field numbers
stable. Private facts and internal message formats belong to the owning service.
Bootstrap SQL lives in `devops/postgres/bootstrap/`. See decision 0008.

API-to-context commands and queries start with `contracts/<context>/messaging/`. Update the
API boundary translation and owner adaptor together; keep generated messages out
of application/domain packages. Preserve command IDs, expected versions and plain
receipt material. A transport request ID matches its reply intent; command receipts use the stable
command ID. Annotate required scalar inputs explicitly and keep future optional
inputs optional. See decisions 0007 and 0009.

HTTP changes start with the OpenAPI sources in `contracts/<context>/http_api/`. Regenerate
service-local bundles with `pnpm generate:http` or `pnpm generate:contracts`.
Find business paths and schemas under `<context>/http_api/`, technical
paths under `services/<service>/http_api/`, and generic HTTP components under
`shared/http_api/`. The [contract guide](contracts/README.md) maps common changes to sources.
Update the relevant handler conformance examples; the Go compositions require
complete method/path coverage and the integration gate validates live owner DTOs.
The generator also rebuilds frontend operation/request/response types and typed
API wire mappings. Update explicit field mappings when public and wire names
diverge. Feature HTTP calls use the generated operation client; `pnpm verify`
checks compilation after incompatible OpenAPI request and response mutations.

Database bootstrap runs as an administrator; applications only check schema
version and checksum. Never add startup migrations or schema ownership to runtime
credentials. The initial schema describes a fresh reference installation.
Once it is shared, preserve it and add a reviewed migration instead of silently
changing the checksum beneath existing data.

## Architecture checks

`scripts/check_architecture.py` rejects outward imports, foreign-context
implementation imports, business dependencies in `foundation`, context imports
outside a service's ownership and implicit domain clocks. Its policy has negative
fixtures in `scripts/test_architecture.py`.

All handwritten source and documentation files must stay below 450 lines. Before
introducing an exception, record the responsibility review and explicitly teach
the checker how to recognise it. Generated bindings, schema descriptors and dependency lockfiles have explicit
reviewed exceptions; their generators own their structure.

Keep these guides and the relevant accepted decision in the same change when a
rule changes intentionally. Do not add staging, production or image-release
workflows to this development reference.


Request bindings live in `services/api/adaptors/messaging/generated/`,
`services/storefront/contracts/requests/generated/`,
`services/operations/src/operations/adaptors/generated/cafe/requests/` and
`services/engagement/src/adaptors/generated/{requests.json,request-types.ts}`.
They are compiler-owned and explicitly classified. Go/Python and generated
TypeScript interfaces carry generated-file headers; descriptor JSON is identified
by its `generated/` directory. A domain service receives only its own request
packages. Generate all service clients with `pnpm generate:contracts`.

For a new event-driven aggregate change, add the owner command DTO and
`CommandHandler`, map the fact in an `IntegrationEventHandler` or
`DomainEventHandler`, and register its durable subscription in the context
composition. Keep the stable subscription identifier in its typed definition and
`adaptors/messaging/subscriptions.json`. Add the private Protobuf command under
that same owner, regenerate with `pnpm generate:contracts`, and exercise duplicate
acceptance plus command execution. Resolve all new providers in the composition
test. `pnpm verify` rejects aggregate mutation outside command execution and DI
imports outside composition. Published class-name changes must never silently
rename stored command identities, queues or wire fields.

Use [the developer workflow](docs/developer-workflow.md) for adding a use case in
each language. `pnpm test:focused <service> --context <context>` gives local
feedback; `pnpm verify` also compares historical contracts. Application handlers
cannot invoke another command handler or execute an aggregate store twice.
Subscription declarations must be exhaustive at composition. See decision 0012.
