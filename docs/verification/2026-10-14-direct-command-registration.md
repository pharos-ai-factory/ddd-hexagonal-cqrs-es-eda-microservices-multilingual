# Direct command-to-handler registration — 14 October 2026

## Scope

All runtime command registrations in Menu, Ordering, Preparation, Collection,
Loyalty and Communication now declare their command name or private codec beside
the plain handler factory. Shared composition helpers supply the executor and
aggregate transaction. Go projection preparation and Communication provider
receipt preparation complete before acquiring the aggregate transaction.

The change preserves command identities, receipt material, expected versions,
published contracts, stored private-message bytes and database schemas. Scaffolds,
context navigation, contributor/agent guidance and architecture negative fixtures
follow [decision 0023](../decisions/0023-direct-command-registration.md).

## Executed checks

- `pnpm dev:doctor`: required tools and Docker engine available.
- `pnpm check --changed`: formatting, types, lint and architectural checks passed.
- Focused lanes passed for `storefront --context menu`, `storefront --context ordering`,
  `operations --context preparation`, `operations --context collection`,
  `engagement --context loyalty` and `engagement --context communication`.
- Registration/composition checks passed in all three languages. They exercise
  concrete handler dispatch, aggregate selection, parsed input and metadata,
  private codec delivery, duplicate registration rejection, projection reads before
  transactions, separate concurrent delivery receipts and provider failure before
  transaction entry. Policy and scaffold regression tests passed.
- `pnpm verify` passed for all five applications: generated drift, historical
  compatibility against `caa3206`, 138 tooling tests, Go race-enabled native tests,
  62 Python tests, 46 Engagement tests, seven web tests and 51 fast Gherkin scenarios.
  The ten live workflow scenarios are dry-run bindings in this deterministic gate;
  their executed evidence comes from the integration run below.
- `pnpm test:integration` passed on the complete rerun: context migrations,
  Go/Python/TypeScript PostgreSQL and RabbitMQ checks, all 31 live API business
  operations, 12 infrastructure Gherkin scenarios, both browser scenarios,
  diagnostics/readiness, provider recovery and in-flight session revocation.
  All 17 Python infrastructure tests passed, including the publication recovery fixture.
- `git diff --check` passed.

The initial integration run found one fixture still calling the removed Python
`accept`, `start` and `complete` executor providers. It now dispatches through the
registered private command codec and published request handlers. Static checks
passed after that correction, followed by the successful complete integration run.

No environment-dependent checks were skipped. The integration runner removed its
disposable containers, network and volumes. Local logs and build outputs remain
outside version control.
