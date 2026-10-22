# Orval and TanStack Query browser client, 22 October 2026

The API OpenAPI document now drives a committed Orval 8.44 client in
`services/web/src/adaptors/generated/http/`. Every operation has one declared
context or technical tag. `tags-split` generates Fetch functions, TanStack Query
v5 hooks/options, split models, a typed operation manifest, MSW 3 handlers, Faker
10 factories and typed fixtures. The repository HTTP generator creates and checks
the complete tree, including added and removed files.

The web composition owns one TanStack `QueryClient`. Explicit reconciliation
uses generated list requests and query keys; realtime messages merge into the
same application cache with the existing revision guard. Automatic refetch and
retry triggers remain disabled. Session calls use generated functions. The
durable-command adaptor uses the generated manifest and types while retaining
raw responses for uncertain-outcome recovery.

Historical OpenAPI bytes remain compatible. The Orval input transformer removes
the browser-managed `Origin` header from generated signatures and normalises
legacy empty `required` arrays in memory. The authoritative document and embedded
Go bundle keep their published shapes.

## Executed evidence

- `pnpm generate:http` and `pnpm check:http` passed. Orval regenerated 141 files
  across six context tags and three technical tags.
- `pnpm check:contracts` regenerated all contracts in an isolated checkout and
  found no changed, missing or unexpected output.
- `pnpm test:focused web` passed formatting, strict TypeScript, ESLint,
  architecture checks and nine native web tests.
- Frontend schema-mutation tests passed for incompatible request/response fields,
  required command headers and query parameters, and compatible optional
  additions. OpenAPI tests also require one declared tag per API operation.
- `pnpm check --changed` passed all changed-file static and documentation gates.
- `pnpm verify` passed historical compatibility, generated drift, 174 script
  tests, Go race/native tests, Python and TypeScript suites, nine web tests and
  54 fast Gherkin scenarios.
- `pnpm test:integration` built all five applications and passed PostgreSQL and
  RabbitMQ recovery, 31 live OpenAPI business operations, 12 infrastructure
  Gherkin scenarios, provider recovery, both browser scenarios and in-flight
  session revocation.

All required environment-dependent checks ran. Disposable containers, networks
and volumes were removed. Docker Compose reported its existing non-Swarm secret
ownership warnings. Remote CI, load testing and manual UAT were not run.
