# Generated identifiers and direct query registration — 14 October 2026

## Scope

Menu, Ordering, Preparation, Collection, Loyalty and Communication use generated
context-specific command, query, integration-event, private-domain-event and
subscription identifiers. Private command identifiers come from their compiled
owner envelopes. Query composition names each identifier, owner parser, concrete
handler and reply mapper together; grouped resource endpoint wrappers are removed.

The change preserves published wire names and field numbers, private stored bytes,
queue and receipt identities, expected versions, transaction ownership and domain
invariants. Query handlers retain their named read repositories, missing-resource
behaviour, revisions and complete/paginated reads. Scaffolds, context navigation,
architecture fixtures and guides follow [decision 0024](../decisions/0024-generated-message-identifiers-and-direct-queries.md).

## Executed checks

- `pnpm dev:doctor`: required tools and Docker engine available.
- `pnpm generate:contracts`: generated request/message identifiers and bindings.
- `pnpm check --changed`: formatting, types, lint and architectural checks passed.
- Focused lanes passed for all six contexts: Storefront Menu/Ordering, Operations
  Preparation/Collection and Engagement Loyalty/Communication.
- Direct query dispatch tests passed in Go, Python and TypeScript. They exercise
  real query handlers, complete lists exceeding 100 items, missing roots, revisions,
  pagination inputs/continuations, invalid page rejection, storage failures and
  duplicate/foreign identifier rejection.
- Typed negative callers reject raw identifiers and mixed message roles. Generator
  tests cover context selection, overlapping owner request names, private event
  scope, manifest-driven subscription additions and private commands absent from
  their compiled owner envelope. Architecture and scaffold regressions passed.
- `pnpm verify` passed for all five applications: generated drift, historical
  compatibility against `caa3206`, 143 tooling tests, Go race-enabled native tests,
  65 Python tests, 48 Engagement tests, seven web tests and 51 fast Gherkin scenarios.
  Its ten live workflow scenarios use dry-run bindings; their executed evidence
  comes from the integration run below.
- `pnpm test:integration` passed: Go/Python/TypeScript PostgreSQL and RabbitMQ
  checks, all 17 Python infrastructure tests, 31 live API business operations,
  12 infrastructure Gherkin scenarios, both browser recovery scenarios, provider
  recovery, diagnostics/readiness and in-flight session revocation. The query
  consumer remains responsive during a blocked command transaction. Lost replies
  recover the original committed outcome on an identical retry.
- `git diff --check` passed.

No environment-dependent checks were skipped. The integration runner removed its
disposable containers, network and volumes. Local logs and build outputs remain
outside version control.
