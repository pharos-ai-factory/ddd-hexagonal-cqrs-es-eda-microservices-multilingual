# Catalogue-scoped broker authority, 14 October 2026

## Change

[Decision 0022](../decisions/0022-catalogue-scoped-broker-authority.md) introduces a
dedicated topology credential and generated exact queue permissions for all six
contexts and the API. The [permission guide](../broker-permissions.md) describes
resource/topic checks, private subscription ownership, API request/reply access
and the topology user's required dead-letter authority.

Generation expands subscription event and durable command queues, retries, dead
letters and request/reply queues. Negative fixtures reject resource collisions
and wildcard consumers. Every managed user receives explicit topic permissions.
Compose mounts only the dedicated topology URL into bootstrap. The launcher
restarts managed broker clients when applying permissions. Existing delivery
recovery fixtures use declared queues with their actual owner credentials.

## Executed evidence

- `pnpm generate:contracts` passed. Full verification reproduced the generated
  permission catalogue without drift and preserved historical wire compatibility.
- `pnpm check --changed` passed after implementation and again after the final
  documentation and verification record.
- Focused Storefront Menu and Ordering tests passed. The broker integration test
  package also compiled before the infrastructure run.
- Eight broker permission tooling tests and six secret-file tests passed.
- `pnpm verify` passed all five applications: 135 Python tooling tests, both Go
  modules with race checks, 60 Operations tests, 44 Engagement tests, seven web
  tests and 51 fast Gherkin scenarios, alongside static and contract gates.
- `pnpm test:integration` passed with real PostgreSQL, RabbitMQ, Valkey and
  Centrifugo. The dedicated topology user declared the complete catalogue and
  repeated private/command declarations successfully. Owner probes covered every
  generated queue, own domain/integration publication and delivery exchanges.
  All six API request/reply paths passed for both request queue kinds.
- RabbitMQ rejected runtime declarations/deletions, queue binding/unbinding,
  exchange binding, foreign queue reads, foreign publications, forged replies,
  unregistered request keys, default-exchange bypasses and unregistered topology
  resources/binding keys.
- Go recovery, 17 Operations infrastructure tests and 22 Engagement infrastructure
  tests passed using owner credentials. Live checks passed all 31 API operations,
  lost-reply recovery, command/query independence, ten workflow scenarios and the
  12 combined infrastructure Gherkin scenarios. Provider recovery and paused
  reward issuance completed through the restricted topology.
- Both browser recovery tests, five Valkey authority tests and the delayed
  authentication/session-revocation test passed. Readiness, diagnostics and
  realtime permission checks passed.
- The disposable Compose containers, network and volumes were removed.

No required environment-dependent check was skipped. The deterministic workflow
binding dry run is followed by execution of those ten scenarios in the live
integration lane. Existing pytest-bdd deprecation and browser container platform
warnings remain. Remote CI and load testing were outside this local verification.

Logs are under `.local/`: `broker-generation.log`, `broker-unit.log`,
`broker-secrets.log`, `broker-focused-menu.log`, `broker-focused-ordering.log`,
`broker-check.log`, `broker-check-final.log`, `broker-verify.log` and
`broker-integration.log`.
