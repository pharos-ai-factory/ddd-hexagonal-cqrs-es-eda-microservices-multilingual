# Executed verification

## Owner HTTP boundary cleanup, 9 October 2026

Retired direct business HTTP routes and JSON decoders across Storefront,
Operations and Engagement. Go composition now injects typed executors and query
handlers directly into messaging registries. Owner HTTP retains health and
authenticated diagnostics. Public API OpenAPI/Protobuf contracts and frontend
bindings remain unchanged; the 17 intentional Storefront development operation
removals have a fixed compatibility exception documented by
[decision 0016](decisions/0016-owner-messaging-boundary.md).

`pnpm verify` passed all five applications, including the new architecture
negative fixtures, native operational surface checks, exhaustive public-path
rejection probes, historical compatibility, Go race/vet, Python/TypeScript checks
and all 51 fast Gherkin scenarios.

`pnpm test:integration` passed database and RabbitMQ recovery, all 31 live API
business operations, all 12 infrastructure scenarios, authenticated business-path
rejection on every owner process, both browser scenarios and session revocation.
All required environment-dependent checks ran. Disposable resources were removed.

Changed-document local links and `git diff --check` pass. Remote CI and load tests
were not run. Local logs: `.local/owner-http-verify.log` and
`.local/owner-http-integration.log`.

## Named repositories and central command execution, 9 October 2026

Decision 0015 names all eight read and write repositories across Go, Python and
TypeScript. All 21 feature command handlers receive named write repositories and
a target-only command context. Central executors own transaction and delivery
bookkeeping. Owner mappers select outgoing aggregate facts during save; Reward
and Notification now record their creation facts inside the aggregate.

`pnpm verify` passed all five applications, generation/compatibility, architectural
checks, DI graphs, Go race/vet, strict Python/TypeScript checks and all 51 fast
Gherkin scenarios. Regression checks cover write capabilities outside commands,
transaction coordination in application code, repository identity/lifetime,
rejection after staging, snapshot isolation and preflight projection reads.

`pnpm test:integration` passed real transaction rollback/replay/concurrency,
PostgreSQL/RabbitMQ recovery, all 31 live OpenAPI operations, all 12 infrastructure
scenarios, both browser scenarios and the separate session-revocation fixture.
All required environment-dependent checks ran, and disposable resources were
removed. The final run and cleanup are recorded in
`.local/hidden-integration-final.log`.

Public schemas, stored message definitions and queue identities retain their
existing contracts. Generated files match their sources. All 182 local links in
the changed guides resolve; command/query scaffolds were inspected in all three
languages and `git diff --check` passed. Remote CI and load tests were not run.

Logs: `.local/hidden-verify-final.log` and `.local/hidden-integration-final.log`.

## Working guidance, 9 October 2026

Agent and contributor guidance now applies decisions 0010–0014 consistently.
The decision index identifies later refinements; extension recipes cover named
commands/queries, durable subscriptions, DI, contract generation and browser tasks.
Transaction diagrams and query terminology match the implemented boundaries.
The documentation follow-up adds no runtime behaviour.

`pnpm verify` passed all five applications, generated/historical contract checks,
architecture and negative fixtures, Go race/vet, strict Python/TypeScript checks
and all 51 fast Gherkin scenarios. `pnpm test:integration` passed persistence and
broker recovery, all 31 live OpenAPI operations, all 12 infrastructure scenarios,
both browser scenarios and the separate in-flight session-revocation fixture.
No required environment-dependent checks were skipped; disposable resources were
removed. Remote CI and load tests were not run.

All 229 local link targets across the 24 changed Markdown files resolve.
`git diff --check` passed. Logs: `.local/docs-guidance-verify.log` and
`.local/docs-guidance-integration.log`.

## Context navigation and query layout, 9 October 2026

Decision 0014 adds 16 named query/handler pairs, application read models and
reader ports across all six contexts. Restoration and private command decoding
live in owning adaptors. Domain modules and event reactions have business-named
files. Browser features follow user tasks; recovery infrastructure remains shared.
The context catalogue, READMEs and role-aware scaffolds provide checked navigation.

`pnpm verify` passed generated/historical contract checks, architecture across
79 Go packages, compiler/AST negative fixtures, both Go race/vet lanes, strict
Python/TypeScript checks and all 51 fast Gherkin scenarios. Read-boundary tests
cover absent resources, revisions, continuation identities, corrupt snapshots,
view isolation and storage failures. Aggregate file moves preserve their original
source bytes. Public schemas and durable identities are unchanged.

`pnpm test:integration` passed all persistence/broker lanes, all 31 live OpenAPI
operations, all 12 infrastructure scenarios, provider/request recovery, readiness,
workflow inspection, both browser scenarios and in-flight session revocation.
No required environment-dependent checks were skipped. Disposable containers,
networks and volumes were removed.

Focused Preparation and Collection lanes passed (12 and 17 tests). Query and
subscription scaffolds were written and inspected for all three languages;
scaffold regression tests and catalogue checks passed after final tooling edits.
Remote CI and load tests were not run.

Logs: `.local/screaming-final-verify.log`, `.local/screaming-final-integration.log`
and `.local/screaming-final-architecture.log`.

## Command module layout, 9 October 2026

Decision 0013 organises all 21 command/handler pairs under their six owning
contexts' `application/commands/` directories. Composition roots, adaptors, event
reactions and tests import the command packages/modules directly. Existing Go
scenario and PostgreSQL fixtures moved beside the command package; native runners
include nested packages. The scaffold creates the same directory structure.

`pnpm verify` passed generation and historical compatibility, architecture checks
across 70 Go packages, both Go race/vet lanes, Python Ruff/strict Mypy, native
tests, TypeScript checks and all 51 fast scenarios. Compiler/AST negative fixtures
reject grouped pairs, missing partners and command declarations outside the
command directory. Go mutation checks cover nested application packages.

`pnpm test:integration` passed persistence and broker recovery, all 31 live HTTP
operations, all 12 infrastructure scenarios, readiness and workflow inspection,
both browser scenarios and in-flight session revocation. No required
environment-dependent checks were skipped. Disposable containers and volumes
were removed. Command scaffolds were also written and inspected for all three
languages. Remote CI and load testing were not run.

Logs: `.local/command-layout-verify.log` and
`.local/command-layout-integration.log`.

## Architecture enforcement and developer workflow, 9 October 2026

All eight review findings are addressed in decision 0012 and the
[developer workflow](developer-workflow.md). `pnpm verify` passed historical
OpenAPI/Protobuf compatibility, generated drift, architecture checks, native
tests, Go race tests/vet, strict Python/TypeScript checks and all 51 fast Gherkin
scenarios. Negative fixtures cover handler chaining, captured execution methods,
repeated aggregate-store calls and incomplete or incompatible subscriptions.
Lifecycle tests exercise partial construction, worker draining, cleanup failures
and Fx failure exit status.

`pnpm test:integration` passed persistence/broker recovery, 15 Python infrastructure
checks, command failure identity logging, all 31 live OpenAPI operations, all
12 infrastructure Gherkin scenarios, workflow inspection, browser recovery and
in-flight session revocation. The new readiness fixture checks a stopped owner
and its missing consumers. Its first run exposed a test assumption about immediate
RabbitMQ management statistics; bounded observation fixed the assertion and the
complete rerun passed. No required environment-dependent checks were skipped.

A subsequent readiness addition checks Centrifugo's health endpoint. Its unit
tests and a disposable live stop/restart probe passed: endpoint loss makes
readiness fail, and restart restores readiness. Disposable stacks and volumes
were removed after both live runs.

Focused Loyalty tests, all nine scaffold variants, workflow identity extraction,
doctor output, formatting and Compose Watch configuration validation passed.
Debugger profiles and the interactive watch session were not exercised. Scaffolds
remain reviewable starters with deliberately failing behaviour tests; developers
supply domain policy and registration. Remote CI and load testing were not run.

Logs: `.local/dx-verify-final.log`, `.local/dx-integration-final.log`,
`.local/dx-realtime-readiness.log` and `.local/dx-focused.log`.

## Dependency injection and durable command reactions, 9 October 2026

Decisions 0010 and 0011 record the framework comparison, explicit composition
boundaries and command-before-aggregate-mutation convention. Awilix, Dependency
Injector and Fx resolve the service graphs. Application constructors retain plain
ports, while context resources have independent ownership and shutdown.

`pnpm verify` passed reproducible generation, architecture checks across 68 Go
packages, Go race tests/vet, Ruff, strict Mypy, TypeScript compilation, native
unit tests and all 51 fast Gherkin scenarios. Compiler/AST negative fixtures
reject aggregate mutation outside command execution, captured method aliases,
misplaced write-port use and DI imports outside composition. An additional Python
regression rejects direct private-state access even from a command handler;
its focused suite and the final architecture check passed. Domain tests and
projection/rehydration paths retain their explicit scope.

DI tests resolve every Python/TypeScript context provider and validate the Go Fx
graphs without infrastructure. Fixed command fixtures cover all seven private
subscriptions and preserve exact Protobuf bytes plus original receipt material.
The application event-handler tests exercise translation separately from command
execution. Handwritten files remain below the responsibility limit.

`pnpm test:integration` passed on the final runtime implementation. Both Python
and TypeScript prove that accepting an event leaves aggregate state unchanged,
a dispatch-recording failure rolls back acceptance, duplicates preserve the
original command, conflicting source hashes fail, abandoned command leases recover
exact bytes, stale lease completion is fenced, four failed executions reach the
dead queue, and replay/duplicate publication commit one aggregate transition.
Runtime roles cannot rewrite the saved command bytes. Owner-local version-three
migrations provide receiving command evidence and mutable dispatch records.

The full suite also passed existing migration/receipt/request-reply tests, all
31 live HTTP contract operations, all 12 infrastructure Gherkin scenarios,
realtime and Valkey authority, workflow diagnostics, browser recovery and the
in-flight session revocation fixture. No environment-dependent checks were
skipped. Disposable infrastructure was removed after completion.

Logs: `.local/commands-verify-final.log` and
`.local/commands-integration-verified.log`. Remote CI and load testing were not
run. The static checks cover the supported typed coding conventions; deliberate
reflection or bypass still requires review.

## Contract review follow-up, 8 October 2026

The four review findings are addressed. API and owner transports now use native
context request/reply envelopes; generated Go interfaces select the owner without
merging payload fields. Engagement's generated wire interfaces are scoped by
context too. A compiler regression gives Menu and Ordering the same command name
and field number, generates the API bindings and compiles both packages together.
Historical request fixtures continue to preserve wire bytes.

API ACL generation reads Protobuf compiler descriptors and required-input options.
Multiline annotations and comments preserve the generated mapping; an explicit
mapping to an unknown wire field fails generation. Frontend mutation checks cover
required headers and query parameters, including union/shared operation callers,
and optional additions. A transport test verifies that typed headers and query
values reach the HTTP request.

Reply leases are eight seconds against the existing thirty-second response
lifetime. Go, Python and TypeScript PostgreSQL fixtures abandon real committed
claims without altering their timestamps, reclaim them before expiry and compare
the original response bytes. The Go fixture also rejects stale completion.

`pnpm verify` passed with 44 script checks, both Go race/vet lanes, strict native
type checks, frontend tests and 51 fast Gherkin scenarios. The isolated session
revocation check passed after adapting its networking for Docker Desktop on macOS:
Centrifugo uses published ports and container-to-host requests use
`host.docker.internal`; Linux retains host networking. Application listeners and
revocation assertions are unchanged.

`pnpm test:integration` passed in full: owner migrations, Go persistence/broker
checks, 14 Python infrastructure checks, TypeScript infrastructure checks, 31
live OpenAPI operations, all 12 infrastructure Gherkin scenarios, permissions,
diagnostics, both browser recovery scenarios and the session-revocation race.
No required environment-dependent check was skipped. Disposable containers and
volumes were removed. Local logs are
`.local/review-fixes-verify.log`, `.local/review-fixes-integration-final.log` and
`.local/review-fixes-revocation.log`.

## Six contract enforcement improvements, 8 October 2026

The original five issues and the frontend enforcement gap are implemented and
tracked in [contract improvements](contract-improvements.md). Decision 0009
records required input annotations, explicit ACL mappings, scoped request
packages, independent command/query consumers, durable command reply intents and
OpenAPI-generated frontend types.

`pnpm generate:contracts` completed. `pnpm verify` passed: generated drift checks,
38 script checks, architecture checks across 66 Go packages, Go race tests/vet,
strict Python/TypeScript checks, native optional-field/zero-value tests and all
51 fast Gherkin scenarios. Real frontend source fails compilation after a
required request or response field is renamed, and remains compatible after an
optional request field is added. Historical fixtures preserve command/query wire
bytes across all six owner packages. Generated drift probes reject added,
removed and changed outputs without modifying the source checkout.

`pnpm test:integration` passed owner migration upgrades, Go persistence/broker
checks, 14 Python infrastructure checks, TypeScript infrastructure checks, all
31 live OpenAPI operations, all 12 infrastructure Gherkin scenarios, permissions,
diagnostics and both browser scenarios. Reply encoding failure rolls back the
root and receipt in all three owner languages; fresh transport requests persist
new exact reply bytes for saved command outcomes. Go fixtures reclaim reply
leases and deny completion by stale publishers and response rewrites.

The live RabbitMQ recovery fixture holds a command advisory lock and proves the
query consumer reads committed state independently. It then removes the reply
binding, observes HTTP uncertainty alongside a committed pending reply intent,
and verifies the command queue has no ready or unacknowledged deliveries.
Restoring the binding completes the original dispatch with unchanged bytes;
identical command retries preserve one receipt, event and root transition.

The overall integration gate failed only at the existing isolated session
revocation fixture: `Fixture Centrifugo did not become ready`, before its API
process starts. Its assertions remain unverified on this Docker Desktop host.
The composition/browser session checks passed. Disposable resources were removed.
No environment-dependent lanes were silently skipped.

Final logs: `.local/six-issues-generation.log`,
`.local/six-issues-final-verify.log` and `.local/six-issues-final-integration.log`.
`git diff --check` and the final architecture check passed after documentation
updates.

## Published contract ownership, 8 October 2026

Published sources now sit under `contracts/<context>/{messaging,realtime,http_api}`.
Messaging contains published commands, queries and integration events. Private
Menu, Loyalty and Communication delivery schemas, fixtures and metadata remain
in their owning service adaptors. PostgreSQL bootstrap SQL moved to
`devops/postgres/bootstrap/` with identical contents and checksums.

Expanding component references in all four OpenAPI bundles produced identical
HTTP boundaries. Comparing the previous and current Protobuf descriptors
confirmed unchanged published request/reply, realtime and integration payload
identities; private payloads were removed from the shared Event envelope and
reserved. Menu and Loyalty fixtures prove compatibility with historical private
bytes. Go, Python and TypeScript decode the shared RewardIssued fixture.
Two successive generations produced identical hashes for 135 generated files.

`pnpm verify` passed, including architecture checks across 55 Go packages,
Go race tests/vet, strict Python/TypeScript checks, all 31 API operation mappings
and 51 fast Gherkin scenarios. The subsequent architecture check passed after
the final documentation updates and reproducible generation.

`pnpm test:integration` passed context migrations, real persistence/broker checks,
request/reply authority probes, all 31 live OpenAPI operations, command reply-loss
recovery, all 12 infrastructure Gherkin scenarios, diagnostics, realtime/Valkey
permissions and both browser scenarios. The private reactions completed through
real RabbitMQ and PostgreSQL; duplicate-delivery evidence retained their payloads
as opaque bytes.

The separate session revocation fixture failed Centrifugo readiness before its
API process started: `Fixture Centrifugo did not become ready`. Its assertions
remain unverified on this Docker Desktop host. The integration gate failed for
that lane; disposable resources were removed.

Logs: `.local/context-owned-contracts-final-verify.log`,
`.local/context-owned-contracts-final-integration.log` and
`.local/context-owned-contracts-reproducible.log`. The previous boundary documents
remain under `.local/owner-contract-layout-baseline/` for semantic comparison.

Earlier executed results are preserved in the
[historical record](verification/2026-10-08-before-contract-ownership.md).
