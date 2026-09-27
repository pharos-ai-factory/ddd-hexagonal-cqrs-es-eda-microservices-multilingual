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
`contracts/events/README.md` for compatibility rules. Regeneration must produce
no unexpected change before merging.

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
