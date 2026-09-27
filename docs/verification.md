# Executed verification

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

The [catalogue](../specifications/README.md) contains **63 passing executable
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
