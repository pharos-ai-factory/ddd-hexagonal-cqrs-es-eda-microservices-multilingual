# Adaptor organisation across all services, 14 October 2026

## Change

The responsibility layout now covers Storefront, Operations, Engagement, API and
Web, including Menu, Ordering, Preparation, Collection, Loyalty and Communication.
[Decision 0021](../decisions/0021-service-and-context-adaptor-layout.md) and the
[adaptor map](../adaptor-layout.md) describe the resulting navigation.

Storefront's shared technical persistence and transport move into `adaptors/`.
Context PostgreSQL implementations share the persistence tree with their owner
migrations; query endpoints and private fact mappings sit beneath messaging.
Engagement separates codecs, broker workers, database outboxes and gateway I/O;
Communication owns its outbound notification provider. Private command codecs
have individual use-case filenames. Python context requests, query endpoints and
incoming subscriptions identify their resource or event. API session storage,
revocation work and gateway disconnects have separate files; web exposes its
operation client and projection codec under adaptors.

Composition, tests, generators, integration lanes, guides and all context maps
follow the moved sources. Negative architecture fixtures enforce responsibility
folders and preserve the separation of shared technical code and owner behaviour.
The Preparation private-event correction from [the preceding change](2026-10-14-operations-adaptors.md)
remains covered.

## Executed evidence

- `pnpm generate:contracts` passed; full verification also passed isolated generated
  drift and historical compatibility checks. Published schemas and private wire
  fixtures retain their bytes.
- `pnpm check --changed` passed after the implementation and enforcement changes.
  The final documentation batch was checked again before hand-off.
- `pnpm test:focused storefront --context menu` and `--context ordering` passed.
- `pnpm test:focused operations --context preparation` and `--context collection`
  passed: 17 and 13 tests respectively.
- `pnpm test:focused engagement --context loyalty` and `--context communication`
  passed with native and behavioural coverage.
- `pnpm test:focused api` and `pnpm test:focused web` passed.
- `pnpm verify` passed for all five applications: 127 Python tooling tests, Go race
  checks, 60 Operations tests, 44 Engagement native tests, seven web tests and
  all 51 fast Gherkin scenarios, alongside static and contract gates.
- `pnpm test:integration` passed with real PostgreSQL, RabbitMQ, Valkey and
  Centrifugo. Coverage included Go persistence/broker checks, 17 Operations and
  22 Engagement infrastructure tests, all 31 live API operations, lost replies,
  command/query independence, ten live workflow scenarios, two browser recovery
  tests and the delayed-authentication session revocation test. Provider recovery
  and the paused reward-issuance journey also passed.
- The disposable Compose containers, volumes and network were removed.
- Changed documentation links resolve; `git diff --check` passed.

An initial verification run caught an overly strict architecture rule for
consumer-owned generated schemas. The rule was corrected to retain that existing
allowed boundary, and the complete verification gate then passed.

No required environment-dependent lane was skipped. The verification-only
workflow dry run checks bindings; those scenarios executed against the live
stack in the integration gate. Existing pytest-bdd deprecation warnings remain.
Remote CI, load testing and deployment were outside this local change.

Logs are under `.local/all-adaptors/`: `generate-all.log`, `focused-*.log`,
`check-rule-fix.log`, `check-handoff.log`, `verify-final.log` and `integration.log`.
