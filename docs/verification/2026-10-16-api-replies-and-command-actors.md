# API reply routing and command actors — 16 October 2026

## Executed checks

- `pnpm verify` passed for all five applications: formatting, static analysis,
  architecture, generated drift, historical compatibility, script tests, Go race
  tests, Python and TypeScript tests, frontend checks and 54 fast Gherkin scenarios.
- `pnpm check --changed` passed after the final recovery-fixture change.
- Docker Go Staticcheck passed for the API using `CAFE_GO_CONTAINER=1`, with
  `STATICCHECK_CACHE=/cache/staticcheck` and the normal UID mapping. Unit probes
  also cover rootless and UID-mapped Docker command construction.
- `pnpm test:integration` ran against the disposable Colima stack. Context
  migration upgrades, real PostgreSQL and RabbitMQ tests in all three runtimes,
  all service images, 31 live OpenAPI operations, application workflows, browser
  recovery, session revocation and permission probes passed.

Two independent API broker clients exchanged concurrent commands and queries
with Menu (Go), Collection (Python) and Loyalty (TypeScript). The fixture sent
24 requests before and 24 after one client reconnected, without lost replies.
Database tests in each runtime confirm that changing the authenticated actor
while retaining command identity produces an idempotency conflict.

The reply-loss test closes the originating API connection after the owner starts
a command. The HTTP response remains uncertain; the owner commits exactly once,
persists the original reply destination and acknowledges independently of delivery.
A retry over the new connection recovers the recorded outcome. The inaccessible
original reply expires without changing its bytes or route.

## Scope

Production deployment, Minikube/k3s qualification and UAT were not run. The
integration harness removed its owned containers and volumes. Apply context
migration version five, provision the generated permissions and upgrade owner
responders before introducing the new API callers. Legacy reply addresses remain
available for rollout traffic.
