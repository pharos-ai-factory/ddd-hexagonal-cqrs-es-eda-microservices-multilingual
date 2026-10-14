# Durable domain-to-integration publication, 14 October 2026

Implemented [decision 0019](../decisions/0019-durable-domain-event-publication.md)
across Go, Python and TypeScript. All six integration-event paths now start with
aggregate-owned domain facts, private outbox delivery and independent publication
subscribers. Repository guards reject direct integration staging on aggregate saves.

Executed successfully:

- `pnpm check --changed`: formatting, strict types, lint, architecture and command
  boundaries across the affected services and tests.
- `pnpm verify`: generated drift, historical compatibility, 121 script tests,
  both Go race suites, 56 Python tests, 44 Engagement tests, seven browser unit
  tests and all 51 fast Gherkin scenarios.
- `pnpm check:compatibility --against 2a5b44c73655d8cd78cfeaf162dc033406b14de5`:
  published interfaces and stored private fixtures match the pre-refactor baseline.
- `pnpm test:integration --env-file .local/event-refactor/integration.env`:
  context migrations, Go PostgreSQL/RabbitMQ, 17 Python and 22 TypeScript component
  tests, all 31 live API operations, all 12 infrastructure Gherkin scenarios,
  request/reply recovery, permissions, readiness, diagnostics, both browser recovery
  scenarios and session revocation.

New database tests fail the final receipt insert after staging outgoing integration
bytes and dispatch, proving rollback in every language. Concurrent redelivery
retains one receipt and the exact original outgoing bytes. Another subscription
processes the same private fact independently. The Python RabbitMQ test closes a
channel after the database commit but before ACK, then checks redelivery and a
separate subscriber's copy. The TypeScript test advances the Reward before
publication and verifies the original fact's revision, causation and occurrence time.

The live concurrent-reward scenario now waits for the publishing subscription's
receipt after observing its Reward: a committed root can legitimately await
integration publication. Native tests retain existing private fixtures and add
fixed fixtures for all six new private publication facts. Manifest negative tests
reject foreign, public-source, mismatched-kind and ambiguous publication bindings.

The disposable stack was removed. No environment-dependent check was skipped.
Hosted CI, load testing, independent human UAT and deployment were not run.
Final logs are under `.local/event-refactor/`: `verify3.log`, `integration2.log`,
`compatibility-current.log` and `check-final.log`.
