# 0003 — One Gherkin catalogue with native service runners

Status: accepted, 27 September 2026.

## Decision

Keep language-neutral executable behaviour in `specifications/<context>` and
`specifications/workflows`. Use Godog for Go, pytest-bdd for Python and Cucumber
for TypeScript. Application bindings call real handlers through explicit test
ports. Shared specifications do not create a shared business implementation.

Every scenario has a stable identity and a named execution boundary. The root
catalogue checker validates syntax, identities and registered owner/lane pairings.
Native runners fail on undefined or pending steps. Reports remain local.

## Evidence boundaries

`@fast` establishes aggregate/application policy using scenario-scoped decision
probes. Those probes are not substitutes for transactional adaptors.

`@postgres` runs Ordering's real command and projection ports while application
workers are stopped in the isolated integration fixture phase. It establishes
stored rejection/retry behaviour and browser publication atomicity.

`@integration` drives the Go API and running services. The harness may observe
owner databases and resend committed public/private wire messages under the
publishing owner's credentials. These are test fixtures, not production
cross-context read permissions. Duplicate scenarios await consumer receipts
before asserting absence of another business effect.

Keep focused unit, persistence, broker and Playwright tests. Gherkin expresses
business examples; it does not require every technical check to become a feature
or every service-outage exercise to become one opaque step.

## Consequences

The deterministic gate includes all fast scenarios and validates workflow step
bindings without claiming live execution. The isolated infrastructure gate runs
the PostgreSQL and live workflow scenarios alongside the existing recovery and
browser evidence. This preserves a readable common specification while exposing
idiomatic implementations in each language.
