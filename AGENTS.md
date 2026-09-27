# Reference implementation rules

Use British English. Read `ARCHITECTURE.md`, `DDD.md`, `CLIENT-SUBSCRIPTIONS.md`
and `TESTING.md` before architectural changes. Follow `CONTRIBUTING.md`.
Accepted decisions live in `docs/decisions/`.

- One command or event-handling transaction changes at most one aggregate.
- Aggregate children have no independent repositories. Reconsider the model
  when immediate business invariants span proposed aggregates.
- Domain and application packages are independent of transport and persistence.
- Domain packages never import another context. Context application packages
  never import another context's implementation.
- Contexts own separate PostgreSQL databases and credentials. Inter-context
  reads use consumer-owned projections or published owner ports.
- Delivered domain events and integration events use the same transactional
  outbox, RabbitMQ publisher confirms, manual acknowledgements, receipts,
  bounded retries, dead-letter and replay mechanisms.
- Persist the aggregate, receipt, outcome and outgoing events atomically.
  Acknowledge incoming messages only after the receiving transaction commits.
- Workflows are eventually consistent. Do not synchronously invoke another
  aggregate's command handler from a command handler.
- Keep domain facts separate from versioned transport envelopes. Domain events
  remain private to their context even when they travel through RabbitMQ.
- Use stable identifiers, immutable published revisions, typed business
  rejections and expected aggregate versions. Never use correlation IDs for
  deduplication.
- Default to current-state persistence. Durable event delivery alone is not
  event sourcing. A change to this choice requires an explicit decision.
- All runtime compositions are development-only. No staging or production
  release workflow is permitted.
- Keep handwritten files below 450 lines. Review and record justified
  exceptions; generated files are explicitly classified.
- Keep contexts inside their owning language service. Do not import another
  service's source or place business policy in the Go API or Next.js.
- Browser updates use server-authorised Centrifugo subscriptions and the separate
  Protobuf projection contract. No browser polling or client-selected channels.
- Persist exact realtime bytes atomically with the root transition. Test recovery,
  stale revisions, independent windows and durable session revocation.
- Run `pnpm verify` and `pnpm test:integration` before hand-off. Report any
  skipped environment-dependent checks.
