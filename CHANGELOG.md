# Changelog

Notable changes to this development reference are recorded here, newest first.

## Unreleased

- Centralise Go API technical errors as typed JSON responses with reusable
  status/code descriptors; preserve session, realtime and owner outcome contracts.
- Organise Go API HTTP handlers into session, realtime, operational and backend
  packages with adjacent tests, a small composition function and internal shared
  transport helpers.
- Define authoritative OpenAPI contracts for every Go HTTP surface, require
  complete route registration, restrict API forwarding to declared operations
  and validate wire schemas in deterministic and live integration checks.
- Add service-local domain and application error types while retaining stable
  recorded rejection codes and retryable infrastructure failures.
- Split the operator workflow into focused panels, validate Engagement snapshots
  at the persistence boundary and simplify command parsing and transaction flow.
- Restrict Centrifugo operations through per-context publisher credentials and a
  separate disconnection credential; remove topology administration from broker
  runtime users.
- Add optional keyset pagination in all three languages and reusable browser
  traversal; unpaginated queries return complete results. Exercise more than
  100 roots and subscription-before-reconciliation recovery.
- Add authenticated workflow diagnostics and redacted worker failure records.
- Add context-owned migration manifests, lifecycle indexes and independent
  checksum/owner verification while retaining current-state persistence.
- Serve statically exported Next.js assets from the ingress and separate durable
  session Valkey from disposable realtime history.
- Mount runtime secrets as protected files and support individual developer
  credential files; atomically repair generated configuration permissions.
- Reject missing/null persisted Go fields before they can become zero prices,
  including nested offers and immutable projections.
- Validate saved receipt identity, version, status and rejection details before
  returning outcomes or acknowledging consumed events.
- Check stored and proposed root identities against their storage keys in all
  three runtimes; corrupt authority rolls back without consuming the attempt.
- Expire in-flight connect grants and retain durable revocation work past their
  validity window; revalidate established connections through the refresh proxy.
- Add PostgreSQL regression tests for missing/null order prices, mismatched root
  identities and malformed command/consumer receipts, including explicit valid
  controls and identical retry after administrative repair.
- Add an isolated Valkey/Centrifugo regression that delays an authenticated
  connection until the real Go session worker has completed logout disconnection.
- Await response handling and session-storage clearance in the browser command
  recovery test; hiding a busy retry button does not establish completion.
- Collect independent integration failures so one failing regression does not
  prevent the other infrastructure lanes from executing.
- Wait for committed consumer projections in the browser journey before adding
  an offer or starting an order; browser publication can precede broker delivery.

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
