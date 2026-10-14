# Static checks and formatting, 14 October 2026

Added `pnpm check`, service/changed-file selection and continuous edit checks.
Verification and CI share strict Operations Mypy, strict TypeScript including
native/Gherkin/browser tests, typed ESLint, Go vet/Staticcheck/errcheck and the
complete handwritten formatting baseline. Focused tests check their owner first;
Compose Watch and editor settings provide feedback during edits.

The checking changes were verified independently of the concurrent event-boundary
work, starting from `dd60d9f`. `pnpm verify` passed with 102 script tests,
compiler/linter negative fixtures, generated drift and historical compatibility,
both Go race suites, Python and TypeScript tests and all 51 fast Gherkin scenarios.

The complete integration runner passed using
`python3 scripts/integration.py --env-file .local/checks/integration.env`:
PostgreSQL/RabbitMQ, all 31 live API operations, 12 infrastructure scenarios,
both browser scenarios and session revocation. All environment-dependent checks
ran; the disposable stack was removed. Hosted CI, load tests, manual UAT and
deployment were not run.

Logs in the primary checkout: `.local/checks/commit-{verify,integration,check}.log`.
