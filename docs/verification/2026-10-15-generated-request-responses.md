# Schema-generated command and query responses — 15 October 2026

Implemented [decision 0029](../decisions/0029-schema-generated-command-and-query-responses.md)
across all six contexts and three owner languages.

## Changes verified

- All 22 commands and 16 queries declare their success response in Protobuf.
  Generation produces 31 response message types, command response aliases and
  per-query success/error response types in the owning language.
- Shared command outcomes, rejections, request errors, nested resources, loaded
  resources and collections use generated plain DTOs. Published read-view names
  alias those DTOs. Handwritten copies and mechanical reply mappers were removed.
- Generated query bindings connect the identifier, parser, handler response and
  checked encoder. Go keeps bindings in a separate registration package so owner
  codecs remain usable by persistence fixtures without importing the request worker.
- Missing-resource behaviour, revisions, complete/paginated reads, optional
  presence and numeric bounds are preserved. Executors continue staging exact
  command reply bytes atomically with the committed outcome.
- Historical comparison checks declared request/response associations. The initial
  required-output annotations receive a narrow, reviewed compatibility refinement
  for the nine existing fields described in decision 0029.
- Architecture rules, query scaffolds, context navigation and educational recipes
  now use generated response contracts and registration functions.

## Executed checks

| Check                                                           | Result                                                                                                                                                                    |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pnpm generate:contracts`                                       | Passed; generated sources accompany the schemas                                                                                                                           |
| `pnpm test:focused storefront --context ordering`               | Passed                                                                                                                                                                    |
| `pnpm test:focused operations --context preparation`            | Passed, 22 examples                                                                                                                                                       |
| `pnpm test:focused engagement --context loyalty`                | Passed, including 11 Loyalty scenarios                                                                                                                                    |
| `pnpm check:compatibility --against HEAD`                       | Passed; existing fields and private fixtures preserved                                                                                                                    |
| `pnpm verify`                                                   | Passed across all five applications, including generation drift, historical compatibility, 167 script tests, Go race tests, 73 Python tests and 54 fast Gherkin scenarios |
| Live integration lane (`python3 scripts/integration.py --keep`) | Passed; equivalent to `pnpm test:integration` with the environment retained for inspection                                                                                |
| `git diff --check`                                              | Passed                                                                                                                                                                    |

Generator fixtures exercise response associations, nested/repeated/optional schema
additions and unsupported shapes. Native compiler fixtures reject mismatched query
responses in all three languages. Encoder and registration tests cover optional
empty/absent values, zero values, overflow, ambiguous alternatives, missing roots,
repository failures and pagination.

The live lane passed Go PostgreSQL/RabbitMQ checks, 18 Python infrastructure tests,
23 TypeScript infrastructure tests, all 31 business API operations, ten workflow
scenarios, twelve infrastructure Gherkin scenario reports, reply-loss recovery,
provider recovery, browser behaviour and session revocation. Ordering's paused
private command and Loyalty's paused reward issuance both recovered correctly.

Initial live runs encountered a root-owned RabbitMQ Erlang cookie in the disposable
volume and a PostgreSQL connection timeout during a TypeScript service restart.
The cookie ownership was corrected in that test volume. The final full live run
passed after deterministic verification completed. Application connection settings
were unchanged. The retained disposable project and its volumes were removed.

The deterministic workflow-binding dry-run intentionally skips ten scenarios;
the live lane executed all ten. No environment-dependent lane was omitted.
