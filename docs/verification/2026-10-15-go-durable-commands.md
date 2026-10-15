# Go durable command reactions — 15 October 2026

Implemented the concrete Go `CommandBusPort` adaptor and Ordering's
`OrderCollected` → `RecordCollection` → Order → explicit realtime publication path.
[Decision 0027](../decisions/0027-go-durable-command-reactions.md) records ownership,
compatibility and the permitted pending state. The [creation recipe](../education/concepts/durable-commands.md)
links the schema, codec, bus provider, paired registration and command handler.

## Executed checks

- `pnpm test:focused storefront --context ordering`: passed the service checks,
  owner codec/domain/application tests and native Ordering scenarios.
- `pnpm check --changed`: passed formatting, documentation, type/lint checks and
  architecture enforcement across the affected applications.
- `pnpm verify`: passed generated drift, historical compatibility, tooling and
  architecture regressions, all five applications' native tests and 54 fast
  Gherkin examples.
- `pnpm check:compatibility --against HEAD`: passed the published interfaces and
  stored private fixtures comparison with the preceding implementation.
- `pnpm test:integration`: passed in the disposable Docker Compose project.
  The project and its volumes were removed afterwards.
- `git diff --check`: passed.

## New behaviour and recovery evidence

| Boundary           | Observed result                                                                                                                                                                                                |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Order invariant    | A placed order accepts matching-customer collection once; drafts and mismatched customers reject; repeated facts leave state and revision unchanged                                                            |
| Private codec      | Historical command bytes round-trip deterministically; original receipt material, source hash, causation and derived command identity survive; malformed envelopes reject                                      |
| Event acceptance   | A dispatch insert failure rolls back the receiving record; concurrent duplicate acceptance creates one record and dispatch; changed source evidence/target rejects                                             |
| Immutable work     | A later mapping retains already accepted command bytes; runtime credentials cannot update those bytes                                                                                                          |
| Fenced relay       | Reclaiming an expired lease preserves bytes and advances its generation; stale completion and retry updates fail                                                                                               |
| Command execution  | Realtime encoding failure rolls back the Order and receipt; replay recovers one committed outcome; another source event for the same collection causes no second transition or browser message                 |
| RabbitMQ           | The private command retains its accepted metadata across three bounded retries, dead-letter transfer and explicit replay                                                                                       |
| Composition        | Fx resolves the graph; missing paired consumers, absent relay/command decoders, duplicate subscriptions and incorrect bindings reject before startup                                                           |
| Live pending state | Ordering accepts and publishes the command while `ordering.record-collection.command` is paused; Order stays at revision 3 with collection unrecorded; restart resumes execution and advances it to revision 4 |
| Browser            | Both independent windows display Ordering's collected badge from its explicit realtime publication; the existing no-polling assertion still passes                                                             |
| Compatibility      | Historical Order snapshots default the new optional flag to false; existing status values and Protobuf field numbers remain stable; query/realtime messages add boolean field 7                                |

The integration gate also passed seven migration tests, the Go PostgreSQL/RabbitMQ
race lane, 18 Python and 23 TypeScript infrastructure tests, 31 live API operations,
ten live Gherkin workflows and twelve infrastructure scenario reports. It retained
provider failure recovery, diagnostics, read-only workflow inspection, broker and
realtime authority, Valkey isolation, two browser scenarios and the separate
session-revocation race test.

The paginated OpenAPI compatibility regression now compares disjoint array/object
branches using the ordinary directional field rules. Tests still reject removed
fields/guarantees, incompatible types, response enum expansion and other union changes.
Readiness also recognises a paused private command consumer independently of its
active event-acceptance consumer. Scaffolds use the implemented Go bus and helper.

## Corrections found during verification

The initial infrastructure run exposed a required-field assumption for historical
Order snapshots and a missing Storefront Compose pause setting. The next startup
check exposed readiness's paired-consumer pause assumption. These were corrected
before the successful complete run. The scenario catalogue also caught two reused
identifiers; the new native cases use ORDER_013–ORDER_015.

These are local development results. Remote CI and production deployment were
outside this run; the repository remains development-only. No required local
infrastructure lane was skipped.
