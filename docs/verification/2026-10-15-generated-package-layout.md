# Generated package layout — 15 October 2026

Moved 13 Go files across eight wholly generated packages beneath `generated`
directories. Request/event models, event identifiers, shared response encoders and
owner query registrations use ordinary filenames. Updated generators, imports,
scaffolds, boundary checks, context navigation and source-linked recipes.
[Decision 0030](../decisions/0030-generated-package-layout.md) records the policy.

Compared every relocated file with its pre-move contents: regeneration reproduced
all 13 exactly after normalising import paths. All former file locations are
absent. Mixed packages retain their generated filename markers. The layout change
preserves runtime behaviour, contract fields, wire identities and stored bytes.

Executed successfully:

- `pnpm generate:contracts` and `pnpm format`.
- `pnpm check --changed`, including architecture, types, lint and documentation.
- 42 targeted script tests covering generated placement, architecture, compiler
  generation, events and response generation. Fixtures verify that request wire
  regeneration preserves plain models, stale wire/event outputs are removed,
  handwritten codecs survive and generated paths retain inward dependencies.
- `pnpm test:focused storefront --context menu` and
  `pnpm test:focused storefront --context ordering`.
- `pnpm verify`: isolated generated-output reproduction, historical compatibility,
  171 script tests, both Go modules with race detection, 73 Python tests,
  53 Engagement tests, nine web tests and 54 fast Gherkin scenarios.
- `pnpm test:integration --env-file .local/generated-package-layout/integration.env`:
  context migrations; Go, Python and TypeScript database/broker fixtures;
  durable command and publication recovery; 31 live API operations; ten live
  Gherkin workflows; realtime permissions; diagnostics; browser recovery;
  Valkey permissions and durable session revocation.

The integration runner removed disposable project `cafe-reference-test-683c6ed9`
and its volumes. Its RabbitMQ cookie was initialised with broker ownership before
startup to avoid the known fresh-volume healthcheck race. No environment-dependent
checks were skipped. Local logs are under `.local/generated-package-layout/`.
