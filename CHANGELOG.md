# Changelog

Notable changes to this development reference are recorded here, newest first.

## Unreleased

## 2026-09-27 — Initial implementation

### Added

- Six café bounded contexts across Go Storefront, Python Operations and
  TypeScript Engagement services, with a separate Go API and Next.js operator
  frontend.
- Domain-driven models, hexagonal ports and adaptors, CQRS and current-state
  persistence with separate PostgreSQL databases and credentials per context.
- Atomic aggregate transitions, recorded command outcomes and transactional
  outboxes for private domain events and published integration events.
- RabbitMQ publisher confirms, manual acknowledgements, consumer receipts,
  bounded retries, dead-letter handling and explicit replay.
- Versioned Protobuf event contracts and separate browser projection contracts,
  with generated bindings and Python type declarations.
- Server-authorised Centrifugo subscriptions, atomic realtime publications,
  independent browser windows, history recovery and durable session revocation.
- Persisted browser command identities that survive uncertain responses and
  reauthentication.
- Explicit Python command, event, outcome and snapshot types, immutable domain
  facts, typed aggregate ports and strict Mypy checks for source and tests.
- Executable Gherkin specifications across all six contexts, deterministic
  verification, and real PostgreSQL, RabbitMQ, Valkey and Chromium checks.
- Isolated development Compose environments, a local notification provider
  simulator, and CI verification of the five applications.
- Architecture decisions, modelling guides, contract documentation, contributor
  guidance, upstream attribution and an executed verification record.

### Verification

- `pnpm verify` and `pnpm test:integration` passed with no required checks skipped.
- All 63 executable Gherkin scenarios passed across the fast and infrastructure
  lanes; both Chromium scenarios passed.
- Contract regeneration passed, including the generated Python type declarations.
