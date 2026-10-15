# Explicit bus ports and realtime publication, 15 October 2026

[Decision 0026](../decisions/0026-explicit-bus-ports-and-realtime-publication.md)
records the typed messaging port names and distributed, explicit browser publication.
The [end-to-end recipe](../education/end-to-end-recipe.md) connects command and query
contracts, parsing, handler registration, aggregate persistence, private facts,
integration publication and authorised realtime delivery.

## Implemented paths

- Go Ordering publishes an order refresh from its declared private `OrderPlaced`
  domain-event subscriber. Python Preparation publishes a ticket refresh from its
  declared private `DrinksReady` subscriber. TypeScript Loyalty explicitly consumes
  `RewardIssued` integration events to publish reward refreshes.
- Other command transitions explicitly stage snapshots through their invocation's
  `RealtimeBusPort`. Selected Ordering, Collection and Loyalty handlers can also
  publish a recorded business rejection, including when the target root is absent.
- Command publications commit with their state/outcome. Event publications commit
  exact browser bytes and dispatch with a separate subscriber receipt. Saves and
  integration publication alone produce no automatic browser message.
- Six owner migrations support independent publications at the same revision and
  outcomes without a root. Existing snapshot field identities and published/private
  messaging compatibility remain intact.

## Executed checks

- Focused Ordering, Preparation and Loyalty lanes passed. Preparation includes the
  delayed-event read-revision regression; the full Go and TypeScript native lanes
  also exercise the equivalent handlers.
- `pnpm check --changed` passed types, lint, formatting and architecture checks.
  Documentation checks verified 143 source excerpts, language coverage, links,
  navigation, search and Markdown exports.
- `pnpm verify` passed for all five applications, including generated drift,
  historical compatibility, 144 tooling tests, both Go modules with the race
  detector, Python/TypeScript/browser native tests and 51 fast Gherkin scenarios.
  The ten workflow scenarios are only bound in this deterministic lane; their
  execution passed separately in the integration lane below.
- Python native tests passed: 68 tests. TypeScript Engagement native tests passed:
  50 tests. New cases cover copied publication buffers, closed invocation scopes,
  current read revisions on delayed delivery and missing authoritative views.
- `pnpm test:integration` passed with real disposable infrastructure. This includes
  all seven migration tests; Go transaction/broker checks; 18 Python and 23
  TypeScript infrastructure tests; all 31 live API operations; 12 infrastructure
  Gherkin scenarios; provider recovery; both browser journeys; session-revocation
  and in-flight authentication checks.
- New PostgreSQL tests prove explicit selection, rollback, command retry deduplication,
  subscriber receipt deduplication, conflicting source detection and rootless rejection
  publication. The live journeys exercise all three declared realtime subscriber paths.
- Browser inspection confirmed all eleven end-to-end recipe steps render, including
  source links and the navigation to the related concepts.

The disposable integration containers, volumes and network were removed. No required
environment-dependent checks were skipped. Dependency deprecation and browser-image
platform warnings did not fail the checks. Remote CI and deployment were not run.

## Remaining implementation difference

Go defines `CommandBusPort`, while its current event reactions update projections.
It still lacks a concrete durable private-command adaptor and paired consumer example
for an aggregate-changing reaction. Python and TypeScript have those implementations.
The [coverage page](../education/implementation-coverage.md) and recipes state this gap.

Logs remain under ignored `.local/bus-refactor/`, including `integration-3.log`,
`verify-4.log`, `native-python-2.log`, `native-ts.log`, `check-final.log` and the
focused owner logs. A concurrent documentation-build collision was resolved by
running the final verification alone.
