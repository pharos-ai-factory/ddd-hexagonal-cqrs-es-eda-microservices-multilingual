# Operations adaptor organisation, 14 October 2026

## Change

Preparation completion now stages `preparation.drinks-ready.domain` from the
aggregate's `DrinksReadyDomainEvent`. The existing private subscription publishes
its integration event independently. Before the fix, importing the ticket mapper
failed because it referenced the removed `DrinksReady` type; its old mapping also
attempted direct integration publication during aggregate saving.

Shared Operations adaptors now have persistence, messaging, realtime and
operational HTTP packages. Database leases and reply storage are separate from
network relays, event and realtime encoding have separate modules, and private
command encoding is separate from its receiving outbox. Request envelope factories
are generated from the selected owner schemas. Composition, tests and generator
imports use the concrete relocated modules.

[Decision 0020](../decisions/0020-operations-adaptor-layout.md) and the
[adaptor map](../../services/operations/src/operations/adaptors/README.md) explain
ownership, navigation and transaction/delivery boundaries. Architecture negative
fixtures enforce the new shared package responsibilities. The architecture gate's
Python import reader has its own module to keep the gate below the file-size limit.

## Executed evidence

- `pnpm generate:contracts` passed. Generated changes contain relocated codec
  imports and the new request-envelope catalogue; published schemas and private
  wire fixtures retain their existing bytes.
- `pnpm check --changed` passed after the source reorganisation and again after
  documentation and regression additions. Strict Mypy, Ruff, formatting and
  repository architecture checks passed.
- `pnpm test:focused operations --context preparation` passed: 17 tests, including
  completion through the named handler, private staging and decoding.
- `pnpm test:focused operations --context collection` passed: 13 tests.
- `pnpm verify` passed for all five applications: generated drift, historical
  compatibility against `caa3206`, 123 script tests, Go race checks, 60 Python tests,
  TypeScript/native checks and all 51 fast Gherkin scenarios. Its workflow dry run
  checked bindings; the live scenarios ran in the integration gate below.
- `pnpm test:integration` passed, including all 17 Python infrastructure tests,
  PostgreSQL/RabbitMQ recovery and fencing, 31 live API operations, readiness,
  diagnostics, live Gherkin workflows, provider recovery, browser recovery and
  session revocation. The disposable Compose resources were removed.
- `git diff --check` passed. New adaptor/decision navigation links resolve.

All required environment-dependent checks ran. Existing pytest-bdd deprecation
warnings were reported. Remote CI, load tests, manual UAT and deployment were
outside this local verification.

Logs are under `.local/operations-adaptors/`: `generate.log`, `check.log`,
`preparation.log`, `collection.log`, `verify.log` and `integration.log`.
