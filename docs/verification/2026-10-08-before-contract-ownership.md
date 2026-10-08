# Executed verification

## API command and query transport, 8 October 2026

The API now translates OpenAPI HTTP requests into separate Protobuf commands and
queries over RabbitMQ. Go, Python and TypeScript owner adaptors construct plain
application types and return typed replies. Dedicated broker permissions separate
API request publication, owner request consumption and owner reply publication.
Bootstrap-owned quorum queues and confirms preserve delivery boundaries; command
receipts recover committed outcomes after redelivery or an uncertain HTTP result.

`pnpm verify` passed, including 54-package architecture checks, Go race/vet,
strict Python/TypeScript checks, all 31 API operation mappings and 51 fast Gherkin
scenarios. Native regressions cover zero values, missing fields, foreign contexts,
pagination and malformed replies. Two successive generations preserved identical
hashes for all 134 generated contract files.

The final `pnpm test:integration` run passed persistence/broker checks, migrations,
request/reply authority denial probes, all 31 live OpenAPI operations, all 12
infrastructure Gherkin scenarios, diagnostics, realtime/Valkey permissions and
both browser scenarios. A deliberately unroutable reply left HTTP uncertain after
one committed transition; identical retries recovered one receipt/outcome, one
domain event and the expected realtime publications without duplicating effects.

The separate session revocation fixture failed Centrifugo readiness before its
API process started. Its assertions remain unverified on this Docker Desktop
host; the integration gate correctly failed for that lane. Disposable resources
were removed. Logs: `.local/rabbitmq-requests-final-verify.log`,
`.local/rabbitmq-requests-final-integration.log` and
`.local/rabbitmq-requests-reproducible.log`.

## Protobuf context and message type sources, 7 October 2026

Delivered payload definitions now live in context-owned `domain_events` and
`integration_events` files imported by the existing Event envelope. Browser
snapshot definitions live in context-owned `snapshots` files imported by the
existing Publication envelope. Message names, packages, fields, numbers, oneofs
and Go package paths remain stable; application and domain types remain plain.
Commands and queries retain their HTTP/OpenAPI contract.

The generator compiles every context source, removes stale bindings, qualifies
Python generated imports inside Operations and resolves TypeScript imports in
a deterministic order. The cached Go compiler plugin is reused only when its
reported version matches the pinned version. Two successive generations produced
identical bytes for all 53 generated contract files. Comparing the previous and
regenerated TypeScript descriptors confirmed identical Protobuf wire definitions
for Engagement events, Engagement realtime and browser realtime.

`pnpm verify` passed after the final generation, including the Go golden byte
fixture, Python/TypeScript compatibility examples, race tests, strict type checks
and 51 fast Gherkin scenarios. The final integration run passed persistence and
broker checks, migrations, all 12 infrastructure Gherkin scenarios, all 31 live
OpenAPI business operations, permissions, diagnostics and both browser scenarios.

Initial checks exposed asynchronous descriptor ordering and a test publisher
that resolved new Protobuf imports relative to the envelope file. Both loaders
now use the schema import root and deterministic loading. A transient DNS/registry
timeout interrupted another generation/build attempt; the completed generation
and integration run above supersede those interrupted attempts.

The separate session revocation fixture again failed Centrifugo readiness before
its API process started. Its assertions remain unverified on this Docker Desktop
host, and the integration gate correctly failed for that lane. Disposable
resources were removed. Final logs are `.local/protobuf-layout-stable-verify.log`
and `.local/protobuf-layout-stable-integration.log`; the reproducibility run is
recorded in `.local/protobuf-layout-reproducible.log`.

## Contract source navigation, 7 October 2026

HTTP source fragments now sit beside their context's schemas under
`contracts/http/contexts/<owner>/`. Service-specific operational, session and
Centrifugo callback fragments live under `services/<service>/`; generic schemas,
responses and parameters live under `shared/`. Complete OpenAPI entry documents
remain directly under `contracts/http/`. Developer guides map common changes to
their source files and identify realtime snapshot ownership.

Expanding local component references in the previous and regenerated bundles
produced identical methods, paths, security, parameters and wire shapes for all
four documents. Generated component names now include their source paths.
Regression examples prove that distinct owners can use the same schema name
without substitution, and that ambiguous component names fail generation.

`pnpm verify` passed, including all Go race tests and schema conformance checks.
The integration run passed live validation for all 31 business operations,
persistence/broker checks, migrations, workflows, permissions, diagnostics and
the main browser workflow. The command recovery browser scenario repeated its
previously observed timeout before its first command, while waiting for an
enabled Create drink button after authentication.

The separate session revocation fixture again failed Centrifugo readiness before
its API process started; those assertions remain unverified on this Docker
Desktop host. The integration gate failed with both failures reported. Disposable
resources were removed. Logs are `.local/contract-layout-verify.log` and
`.local/contract-layout-integration.log`; the previous bundles used for semantic
comparison remain under `.local/http-layout-baseline/`.

## Go API HTTP packages, 6 October 2026

The HTTP adaptor now groups handlers and adjacent tests in `operational`,
`session`, `realtime` and `backend` packages. Each receives its own configuration;
the parent composes the routes and retains complete OpenAPI conformance checks.
Shared security, JSON/error responses and test fixtures live under `http/internal`.
The published routes, credentials, session cookies, error bodies and realtime
authorisation deadlines retain their existing behaviour.

`pnpm verify` passed, including architecture checks across 48 Go packages,
race-enabled tests, vet and all 51 fast Gherkin scenarios. The integration rerun
passed live OpenAPI validation for all 31 business operations, persistence/broker
checks, workflows, permissions, diagnostics and both Chromium scenarios.

The initial browser run timed out before sending its command because the
connection remained in reconnecting state after successful authentication.
Both browser scenarios passed on a fresh full integration run; no browser or
session-worker code was changed. The first trace is preserved in
`.local/api-http-packages-browser-failure.zip`.

Both integration runs failed the separate session revocation fixture's Centrifugo
readiness check before its API process started. Those assertions remain unverified
on this Docker Desktop host. Disposable resources were removed after each run.
Logs are `.local/api-http-packages-verify.log`,
`.local/api-http-packages-integration.log` and
`.local/api-http-packages-integration-rerun.log`.

## Go API technical error responses, 6 October 2026

All seven API technical JSON errors use a shared `ErrorResponse` and reusable
status/code descriptors. The published bodies retain their existing fields;
session/Centrifugo envelopes and forwarded owner outcomes retain their contracts.
Boundary tests cover failed session creation, authentication and revocation,
upstream failure, redaction, cookie/cache behaviour and refusal to forward when
session authority is unavailable.

`BenchmarkAPIErrorResponseEncoding` compared the previous map body with the typed
writer on Go 1.27.1, macOS ARM64, Apple M3 Max. Across three runs the median was
390.7 ns/op for maps and 197.0 ns/op for structs. Allocations fell from 7 to 4 per
response, and allocated bytes from 408 to 96. Both variants use the same JSON
encoder and headers with a body-discarding writer. This measures encoding only;
it makes no claim about overall request latency or throughput. Results are in
`.local/api-errors-bench.log`.

`pnpm verify` passed. The integration run passed the live OpenAPI checks for all
31 business operations and every other lane except the separate session
revocation fixture. Its Centrifugo health endpoint remained unreachable before
the fixture API started, so those assertions remain unverified on Docker Desktop.
Disposable resources were removed. Gate logs are `.local/api-errors-verify.log`
and `.local/api-errors-integration.log`.

## Go API route organisation, 6 October 2026

The API composition now mounts focused session, realtime, operational and backend
route groups. Named handlers and their existing tests follow those responsibilities;
shared security and JSON response helpers have separate files. OpenAPI documents
and externally visible behaviour retain their existing contracts.

`pnpm verify` passed after the refactor. The integration run passed live OpenAPI
validation for all 31 business operations, persistence/broker checks, workflows,
permissions, diagnostics and both browser scenarios. The separate session
revocation fixture again failed to reach its Centrifugo health endpoint before
its API process started, so its assertions remain unverified on this Docker
Desktop host. Disposable resources were removed. Logs are
`.local/api-routes-verify.log` and `.local/api-routes-integration.log`.

## OpenAPI HTTP authority, 6 October 2026

`pnpm verify` passes with authoritative OpenAPI documents, reproducible embedded
bundles and complete route registration checks enabled. Conformance tests cover
all 38 Go API operations, 19 direct Storefront operations, three realtime gateway
operations and three development provider operations. Negative examples reject
malformed requests, missing/wrong response fields and undeclared status codes.

`pnpm test:integration` executed the PostgreSQL/RabbitMQ checks, recovery journey,
12 infrastructure Gherkin scenarios, realtime/Valkey permission probes,
diagnostics and both browser scenarios successfully. The new live OpenAPI probe
passed for all 31 business operations across all six contexts, including owner
DTOs, cursor traversal, 404/409/422 outcomes and missing-version responses.

The integration gate failed during the separate session revocation fixture's
readiness check. On this macOS host with Docker Desktop, its Centrifugo container
started with the memory engine and logged both HTTP listeners, but the host could
not reach its health endpoint through `--network host`. A focused rerun reproduced
the same failure before the fixture API started. The revocation assertions did
not execute; they remain unverified in this run. Disposable Compose resources and
both fixture containers were removed. No other integration lane was omitted.

Logs are `.local/openapi-verify.log`, `.local/openapi-integration.log` and
`.local/openapi-revocation-debug.log`. The existing development dataset was
preserved. Remote CI remains outside these local results.

## Runtime authority and complete queries, 29 September 2026

`pnpm verify` and `pnpm test:integration` pass after the runtime, secret delivery,
query and migration changes in decision 0004. No environment-dependent check was
omitted. All 51 fast and 12 infrastructure Gherkin scenarios passed; the workflow
binding dry run remains separate from the executed infrastructure scenarios.

| Change | Executed evidence |
| --- | --- |
| Optional pagination | More than 100 roots in each language; complete ordinary arrays and keyset traversal without omissions or duplicates; invalid cursor/limit cases and ordinary query helpers |
| Browser reconciliation | Both Chromium scenarios pass with 105 seeded drinks, independent windows, all subscriptions attached before page traversal, live revision guards, history recovery and no business polling |
| Realtime authority | Gateway unit tests preserve original publication bytes; live context credentials cannot publish foreign channels, disconnect sessions or invoke administration |
| RabbitMQ authority | All six runtime users can publish/consume/acknowledge but cannot declare/delete topology or publish another context's private events; recovery/replay suites pass |
| Valkey separation | Five permission checks pass, including rejected realtime authentication to the separate session instance; delayed-connect revocation regression passes |
| Context migrations | Seven real PostgreSQL migration tests cover additive upgrade, retained state, independent versions, rollback and authority checks; runtime rejects schema identity and ledger drift |
| Diagnostics | Four live authenticated endpoints expose owned backlog/revocation metrics; exact backlog increments, redaction, pre-claim failure retention and unavailable authority are tested |
| Secret delivery | File-input ambiguity/error tests and atomic private configuration tests pass; real Compose services start with protected secret mounts, and the existing development env file is mode `0600` |
| Static frontend | The final ingress image serves the Next.js static export; full browser workflows pass without a frontend Node server |

The final local logs are `.local/hardening-verify.log` and
`.local/hardening-integration.log`. Disposable infrastructure and revocation
fixture containers were removed. Existing development business data was preserved.

This remains a development composition. Contexts sharing a language process share
its security boundary; per-context credentials are enforced by each database and
gateway login. OpenBao deployment, live leased-secret renewal, production identity,
retention, load testing and clustered availability remain outside this evidence.
Earlier entries below describe the repository at their recorded dates.

## Regression fixes, 27 September 2026

Both `pnpm verify` and `pnpm test:integration` pass with the corruption and
in-flight authentication regressions enabled. No environment-dependent check
was omitted. The browser specifications also passed a focused strict TypeScript
check.

| Correction | Executed evidence |
| --- | --- |
| Required stored Go fields | Missing/null order prices fail without changing state, receipts or either outbox; explicit zero and paid prices work, and the original command succeeds after repair |
| Root identity | Go, Python and TypeScript reject mismatched stored identities and proposed changes; corrupt queries fail and administrative repair preserves the original attempt |
| Saved receipts | All six command/consumer corruption cases pass in both Go and TypeScript; repair returns the original outcome without rerunning the decision or duplicating intent |
| Pending connection revocation | The real worker completes its disconnect before the held response is released; Centrifugo rejects the expired grant and the revoked session receives no publication |
| Established connections | The healthy control revalidates its original session through the refresh proxy and receives a publication |
| Browser prerequisites | The host waits for committed Drink/Menu projections before dependent commands, preserving all browser request-count assertions |

The Go PostgreSQL/RabbitMQ suites, all 12 Python infrastructure tests, all
TypeScript infrastructure tests, two PostgreSQL and ten live Gherkin scenarios,
the outage/provider recovery journey, both café browser scenarios, four Valkey
permission checks and the separate revocation browser fixture passed. The
deterministic gate includes all 51 fast Gherkin scenarios and the existing
workflow binding dry run; that dry run is not a skipped infrastructure lane.

Local logs are `.local/fix-verify.log` and `.local/fix-integration.log`.
Disposable Compose resources and the separate revocation fixture containers
were removed. The development dataset was preserved. Remote CI, timed one-hour
session expiry, multi-process revocation fault injection and the other explicit
limits below remain outside this verification.

## Captured regressions before fixes, 27 September 2026

The regression additions deliberately assert the required behaviour while leaving
the application defects in place. `pnpm verify` passes. `pnpm test:integration`
executes every infrastructure lane and fails on the newly captured defects:

| Regression | Observed result |
| --- | --- |
| Go missing/null stored order prices | Both cases fail: placement commits a zero price, state, receipts and outgoing publications |
| Go corrupt command/consumer receipts | Six cases fail: empty, null and incomplete rejection outcomes are accepted; command outcomes can be copied into consumer receipts |
| Python mismatched Pickup identity | Fails because collection succeeds for a snapshot identifying another root |
| TypeScript corrupt command/consumer receipts | Five cases fail; the null command outcome already rolls back through the database constraint when copied into a consumer receipt |
| Logout during connection authentication | Fails after the real Go worker completes disconnection: releasing the captured response permits a publication to reach the revoked session |

Explicit zero/paid prices, a matching Pickup identity and the healthy realtime
connection control pass. Repair-and-retry assertions are included after corruption
detection; those assertions remain unexecuted in the failing cases until the
application guards are fixed. No expected-failure or skip markers hide the defects.

The existing PostgreSQL/broker checks, recovery journey, two PostgreSQL and ten live
Gherkin scenarios, both existing browser scenarios and four Valkey permission
checks pass. The browser authentication-recovery scenario now waits for its
successful retry response and session-storage clearance. The revocation fixture
also rejects a cancelled proxy request as evidence, preventing an authentication
timeout from being mistaken for successful revocation.

No environment-dependent lane was omitted. Disposable Compose resources and the
separate revocation fixture containers were removed after execution. Local logs
are `.local/regression-verify.log` and `.local/regression-integration.log`; the
revocation browser trace is under `test-results/session-revocation`.

The remaining sections record the earlier baseline, before these regression
assertions were introduced.

## Previously recorded baseline

Verified locally on **27 September 2026**, Linux ARM64 with Docker/Compose,
Go 1.27.1, Node.js 24, pnpm 10.34.5 and uv 0.12.19. Operations runs with Python
3.14 in its container; the local Python test environment uses 3.13. The pinned
Playwright 1.63.0 container supplies Chromium and its system libraries.

The implementation contains three domain languages, a separate Go API and Next.js.
Remote GitHub CI has been configured but has not run on a remote runner here.

Both required gates pass, including regression coverage for real
PostgreSQL/RabbitMQ failures, Chromium authentication recovery and Valkey
permission checks. No environment-dependent check in either gate was omitted.

| Command / check | Executed result |
| --- | --- |
| `pnpm verify` | Passed service/context dependency rules, relative-import negative fixtures, large-file checks, formatting, both Go modules' race tests/vet, Python Ruff/strict Mypy/tests, Engagement and web type checks/tests, all 51 fast Gherkin scenarios and catalogue/report checks |
| `pnpm test:integration` | Passed the isolated PostgreSQL/RabbitMQ/Valkey/Centrifugo stack, all application image builds, two PostgreSQL and ten live Gherkin scenarios, recovery journey, both Chromium scenarios and four Valkey permission checks |
| `pnpm generate:contracts` | Passed with new Python `.pyi` declarations; existing generated bindings, descriptors, fixture bytes and migration checksum metadata were unchanged |
| Native Gherkin runners | Passed 23 Godog, 11 pytest-bdd and 17 Cucumber fast scenarios through `pnpm verify` |
| Browser layout | Chromium passed the horizontal-overflow assertion at a 390px mobile viewport |

## Python type boundaries

Operations passed pinned Mypy 2.3.1 in strict mode across all 34 handwritten
source and test files. Commands, published event payloads, outcomes and restored
snapshots have explicit field types; aggregate lifecycles use literal statuses
and domain facts use immutable context-owned dataclasses. Command and query
ports carry their aggregate's snapshot type. Generated Protobuf declarations
remain inside the adaptors.

All 36 fast Python tests passed, including a valid typed caller and five rejected
caller examples: an incorrect command field type, a missing command field, a
port for another aggregate, a missing event field and an invalid lifecycle
status. Runtime tests independently reject malformed HTTP input and corrupt
stored outcomes and snapshots. The nine Python PostgreSQL/RabbitMQ checks also
passed against real infrastructure.

## Executable Gherkin evidence

The [catalogue](../../specifications/README.md) contains **63 passing executable
scenarios**, including Scenario Outline example rows: 51 fast, two PostgreSQL
and ten live workflows. Native Cucumber JSON reports are in `.local/bdd/`;
the gate checks each scenario identity and its expanded example count against
the parsed features. No business scenario was skipped. The separately named
workflow dry-run report records binding validation only.

The fast scenarios invoke actual application handlers. They exercise empty menu
and order rejection, immutable published terms, child identity and combined
quantity limits, preparation/collection transitions, exact event selection,
three-credit grant accounting, separate reward issuance, the exact expiry
deadline, single redemption, notification intent and provider failure. Menu
examples additionally preserve an 80-character Greek name on publication and
allow an explicitly free offer.

The PostgreSQL scenarios prove a recorded missing-menu rejection survives
projection arrival and identical retry. A new command succeeds, while conflicting
reuse of a command identity cannot change the original order or browser intent.

The live scenarios resend OrderCollected, RewardEarned, OrderPlaced, DrinksReady
and PickupOpened with new envelope identities. Each waits for its consumer's
committed receipt before checking the original aggregate, outgoing events and
provider effects. Additional examples prove six concurrent pickup collections
converge on six credits/two rewards; competing redemptions yield one success and
one version conflict; customer balances remain independent; a lost provider
response produces one acceptance; and a recorded completion rejection remains
stable after preparation starts.

pytest-bdd's transitive Gherkin parser currently emits Python deprecation warnings
about a positional `maxsplit` argument. These warnings are visible and do not
represent skipped or failed scenarios.

## Persistence and event delivery

Go PostgreSQL tests proved atomic state, receipts and event outbox; terminated
connections before commit; concurrent expected-version checks; conflicting
consumer identities across targets; restricted runtime roles; immutable source
evidence and expired dispatch fencing. Its realtime test separately proved
encoder-failure rollback, singular intent on retry and fenced completion.

Python and TypeScript independently proved real PostgreSQL rollback after a
state write and encoder failure, stable recorded retries, conflicting command
input, concurrent stale versions, atomic browser snapshots and realtime lease
fencing. These are their actual adaptors, not a shared in-memory implementation.

The connection-recovery tests terminate an idle Engagement connection and
prove that the process survives and the pool replaces it. Corrupt Pickup and
Reward snapshots fail without recording command or consumer receipts; after
repair, the original command identity succeeds. Restoration failures remain
server failures at the Python HTTP boundary. Notification content is validated
before provider I/O, with an application probe establishing zero requests for
invalid stored content.

RabbitMQ tests covered mandatory unroutable publication, a connection closed
after consumer commit but before ACK, duplicate-effect prevention, delayed
retries, dead-letter preservation, explicit replay and denial of foreign
private-domain bindings.

The Python order-delivery fixture proved that application payloads receive
integer prices while command receipt input retains the original Protobuf JSON
representation. The consumer's derived command identity, original wire hash and
receipt fingerprint remain unchanged by the typed conversion.

All three runtimes also quarantine invalid retry counters, including negative,
overflowing, non-numeric, fractional, boolean and null values. Quarantine
preserves the original event identity and bytes and uses encodable bounded
retry metadata. Go and TypeScript probes also prove that valid event bytes with
invalid counters never reach routing or handlers; Python proves that the next
valid delivery can proceed.

The cross-language journey stopped Operations while placing an order, resumed
it, collected three orders, observed an earned grant while the private Reward
consumer was paused, then resumed issuance. Four notifications reached the
idempotent local provider despite a deliberately lost acceptance response.
Malformed query identifiers returned the same 400 outcome across the three
language services through the Go API.

Disposable integration data was removed after successful verification; the
development dataset was preserved.

## Browser delivery and access

The real two-window Chromium scenario performed menu publication, order
placement, preparation, collection, Loyalty credit and notification delivery.
It lost a command response after server acceptance and retried with the same
persisted command key, creating one result.

The additional authentication-recovery scenario loses a committed CreateDrink
response, removes the session cookie, receives 401 on retry and signs in again.
All three requests use the same command identity. Session storage retains the
attempt through authentication loss and clears only after recovery; exactly
one drink exists afterwards.

Each window made eight initial owner queries. The subsequent workflow and a
recoverable disconnect added no business GETs. A deliberate missing-history gap
caused exactly one additional eight-query reconciliation in each window.
Logout in one window invalidated access in the other via durable Centrifugo
disconnection. Typed projections and all three language producers were exercised.

Unit regression evidence covers stale/duplicate root revisions, a new history
gap while an older query is in flight, and closing a window with queued work.
API tests cover origin rejection, authenticated connect proxy, server-selected
channels, private cookies and replacement of browser credentials at the service
boundary.

The realtime Valkey login permits explicit connection, script, Pub/Sub and
stream-history commands. Real permission tests deny ACL administration, direct
session-key writes and scripted session-key access, while allowing its own
history hashes. Both Chromium scenarios pass with these restrictions, including
history recovery and logout propagation.

## Explicit limits

No staging/production deployment, real email, load test, multi-node availability,
independent human UAT or remote CI execution was performed. The browser uses a
fixed demonstration operator without a production identity-provider integration.
One-hour session expiry and multi-process revocation fault injection have not
received a separate timed/fault-injection lane; logout propagation is exercised.

Current-state persistence is authoritative. Event-sourced rehydration is not
implemented. Diagnostic lists remain bounded to 100 rows per kind. Cross-platform
Docker networking and the complete container-only Go test lane are unverified;
the required gates used the local Go toolchain. The standalone development demo
was not rerun in this verification pass.

Source paths in this historical record describe the layout at each run.
