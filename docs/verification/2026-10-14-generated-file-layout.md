# Generated-file locations — 14 October 2026

Generated outputs must include `generated` in the filename or sit beneath a folder
named exactly `generated`. Existing `*_generated.go` and `generated.json` paths
remain valid. Dependency lockfiles and Next.js' required `next-env.d.ts` retain
their standard names.

Moved six Python/TypeScript event-contract files and three Storefront migration
assets into `generated/` folders. All nine files are byte-for-byte identical to
their previous contents. Updated generators, imports, embedded-resource paths,
context navigation and decision 0017. Plain DTOs retain their application-facing
role; transport and domain dependency rules remain enforced.

`pnpm generate:contracts`, the four focused context lanes (Preparation, Collection,
Loyalty and Communication) and `pnpm check --changed` passed. `pnpm verify` passed
all five applications, 119 tooling tests, generation and historical compatibility,
Go race tests, 51 Python tests, native TypeScript/web tests and 51 fast Gherkin
scenarios. Placement tests cover accepted filenames/folders, unmarked outputs,
generated headers outside the inventory, ordinary handwritten comments, exact
standard-name exceptions, stale producer removal and unchanged migration bytes.

The full integration runner passed through
`python3 scripts/integration.py --env-file .local/generated-layout/integration.env`.
It covered PostgreSQL/RabbitMQ, migration checks, the café recovery journey, all
31 live API operations, 12 infrastructure scenarios, both browser scenarios and
session revocation. All required environment-dependent checks ran; the disposable
stack and volumes were removed. Hosted CI, load tests and deployment were not run.

Logs: `.local/generated-layout/{check,focused-operations,focused-engagement,verify,integration}.log`.
