# Verification

`pnpm verify` is the deterministic gate for every application:

- Service/context ownership and inward imports across Go, Python and TypeScript;
  policy negative fixtures and large-file responsibility checks.
- Formatting, race-enabled unit tests and vet for both Go modules.
- Reproducible OpenAPI bundles, complete Go route registration and schema
  conformance for every Go HTTP operation; negative wire-drift examples.
- Frozen Python dependencies, Ruff, strict Mypy over handwritten source/tests,
  type-gate negative examples and domain/contract tests.
- Engagement and web TypeScript checks and behaviour tests.
- Owner HTTP boundary guards, operational endpoint authentication and rejection of
  every public business path on each owner service.
- 51 executable Gherkin scenarios through native Go/Python/TypeScript handlers;
  catalogue validation, typed step bindings and a dry run of workflow bindings.

`pnpm test:bdd` runs only the fast Gherkin scenarios. The
[specification catalogue](specifications/README.md) explains tags, ownership,
individual scenario selection and reports under `.local/bdd/`.
The dry run binds workflow steps without executing them and is not evidence
that the live workflows passed.

## Choose the right feedback loop

| Check | Command | What failure means |
| --- | --- | --- |
| One context's native tests and scenarios | `pnpm test:focused <service> --context <context>` | Owned behaviour or its bindings failed; this lane omits full architecture/type/infrastructure checks |
| Every service's native tests | `pnpm test:unit` | A native test failed; standalone Engagement Gherkin scenarios are covered by `pnpm test:bdd`/`pnpm verify` |
| Imports, Python use-case rules, navigation and file size | `pnpm check:architecture` | Ownership, an inward dependency, Python mutation/layout, context navigation or responsibility limit was violated |
| TypeScript mutation and use-case layout | `node scripts/check_command_boundaries.mjs` | A compiler-resolved aggregate/store capability or command/query/reaction layout was violated |
| Go mutation and use-case layout | `python3 scripts/go.py test ./tests/architecture` | A typed Go capability or application package/file layout was violated |
| Generated HTTP outputs | `pnpm check:http` | Embedded bundles, frontend types or API mappings differ from OpenAPI sources |
| All generated outputs | `pnpm check:contracts` | A generated file changed, disappeared or needs adding |
| Historical public/private compatibility | `pnpm check:compatibility --against <commit>` | A published interface or accepted stored-message fixture is incompatible with that revision |
| Complete deterministic gate | `pnpm verify` | At least one generation, compatibility, architecture, type, native or fast-scenario check failed |
| Real infrastructure and browser | `pnpm test:integration` | A real authority, transaction, delivery, workflow or browser boundary failed |

The compatibility gate permits only the 17 direct Storefront business operation
removals recorded by [decision 0016](docs/decisions/0016-owner-messaging-boundary.md).
Public API operations and owner operational endpoints retain full compatibility checks.

Run commands from the repository root unless a service directory is specified.
The Go wrapper defaults to Storefront; `CAFE_GO_PROJECT=services/api` selects the
API module. For focused type checks, use `uv run --frozen mypy` inside
`services/operations`, or `pnpm --filter @cafe/engagement type-check` and
`pnpm --filter @cafe/web type-check` at the root.

Python focused discovery includes `tests/contexts/<context>/` recursively and
`tests/bdd/test_<context>.py`. Go discovers nested packages and their Gherkin
bindings; TypeScript discovers adjacent native tests and runs its tagged context
scenarios. Service-wide transport/infrastructure tests remain outside the focused
context lane. Check new tests land in the appropriate discovered tree.

The three architecture commands above provide focused feedback. `pnpm verify`
runs all three and their negative fixtures. A generated-drift pass proves freshness;
historical compatibility separately checks existing consumers and queued work.
Both full hand-off gates are required, and their observed results belong in
[executed verification](docs/verification.md).

## Real infrastructure evidence

`pnpm test:integration` provisions a disposable Compose project with random ports
and separate credentials. It builds all applications, including the Next.js
static export served by the ingress, and runs:

| Boundary | Required evidence |
| --- | --- |
| Go PostgreSQL | State/receipt/outbox atomicity, terminated connection, concurrent commands, conflicting consumer identities, restricted roles, immutable evidence and fenced leases |
| Python and TypeScript PostgreSQL | Repository scope/lifetime, rejected staged saves, encoder-failure rollback, recorded retry outcomes, conflicting input, expected-version races, atomic browser publications and realtime lease fencing |
| RabbitMQ | Protobuf metadata, confirmed mandatory publication, commit-before-ACK redelivery, bounded retry, dead-letter preservation, replay, private-topic restrictions and denial of runtime topology mutation |
| Context migrations | Independent owner ledgers, additive upgrade preserving roots, idempotent reapplication, checksum/owner rejection and rollback |
| Query completeness | More than 100 roots in each language; complete unpaginated arrays and cursor traversal without omissions or duplicates |
| Cross-language workflow | Menu → order → preparation → collection → account → reward → notifications; Operations outage and paused private Reward consumer |
| RabbitMQ requests | Restricted API/owner authority, commands and queries across all three runtimes, independent command/query consumers, exact committed reply recovery after an unroutable publication, atomic rollback and stable receipts/outgoing intent |
| HTTP contracts | Authenticated probes reject direct business HTTP on all owner processes; live OpenAPI response validation through the Go API for all six contexts, complete/paginated reads, item DTOs, missing versions, missing roots and conflicting command identities |
| Gherkin PostgreSQL | Recorded `menu_pending` after projection arrival, new-attempt success and conflicting key reuse; atomic browser intent |
| Gherkin live workflows | Ten focused scenarios: repeated business facts with new event IDs, duplicate grants/notifications, concurrent collections/redemptions, customer isolation, lost provider response and recorded completion rejection |
| Provider | Acceptance followed by a lost response creates one provider effect |
| Browser / Centrifugo | Two independent windows, more than 100 drinks, persisted uncertain command retry across authentication loss, typed live updates, no business polling, recovered history, missing-history reconciliation, mobile layout and logout |
| Realtime authority | Context credentials cannot publish another channel, disconnect sessions or use administrative APIs |
| Workflow diagnostics | Authenticated owned backlog metrics in all runtimes, pending-work age, redacted pre-claim failures and session revocation counters |
| Valkey authority | Realtime credentials cannot authenticate to session Valkey, write session keys or administer ACL users; permission probes require a disposable project |
| In-flight authentication | A held successful connect response cannot restore access after the real Go session worker has completed logout disconnection |

Fixtures use real PostgreSQL and RabbitMQ. Missing infrastructure is a failure,
not a skipped pass. Component fixtures are cleared only in the guarded disposable
project before the acceptance journey; the development dataset is preserved.
The test project is removed at completion, including on failure. Use
`python3 scripts/integration.py --keep` to retain it for diagnosis.

Independent test lanes continue after assertion failures and the runner reports
their combined failure at the end. Provisioning and journey prerequisites still
stop execution if unavailable. Continuing a lane never turns a failure into a pass.

The corruption regressions use administrator credentials only inside a guarded
disposable project. Missing/null Go prices must fail while explicit zero remains
valid; snapshots in all three runtimes must identify their storage keys and
proposed changes must retain that identity; Go and TypeScript must
reject malformed command and consumer outcomes. Failed attempts leave state,
receipts and outgoing intent unchanged. Administrative repair preserves the
logical root version so the exact original attempt can recover afterwards.

`tests/infrastructure/session_revocation.py` starts separate temporary Valkey and
Centrifugo instances and a host using the actual Go API handler, session adaptor
and disconnect worker. It holds one successful authentication response, observes
completed disconnect work, then releases the response and tests access. The healthy
connection control proves that the fixture can refresh its session and receive
publications. An expired connect grant is an explicit protocol refusal even when
the unauthenticated WebSocket remains open; a cancelled proxy call is not proof
of revocation. The fixture's proxy timeout accommodates the deliberate hold.
On Docker Desktop for macOS, the fixture publishes Centrifugo ports and uses
`host.docker.internal` for container-to-host access; Linux uses host networking.
These test controls are absent from the application entry point. The fixture is cleaned up
on failure; its browser trace remains under `test-results/session-revocation`.

Gherkin duplicate scenarios wait for the new message's committed consumer receipt
before asserting no second effect. Cross-service fixtures use only the Go API,
published wire contracts and explicit database evidence; they do not import
another service's domain code. Fresh identities isolate scenarios. Their shared
provider failure control requires serial execution in one disposable project.

`pnpm test:browser` runs the browser scenario against the existing development
stack. It adds demonstration data and removes test channel history during the
recovery exercise. `CAFE_ENV_FILE` selects a different repository-local env file.
Before adding an offer or opening an order, the test host waits for the owning
consumer's committed published projection using its runtime read credentials.
The browser's menu update alone does not prove RabbitMQ delivery has completed.
This bounded infrastructure observation adds no browser business requests.
Chromium runs in the pinned Playwright Docker image, including its system
libraries; the wrapper handles rootless Docker ownership. It runs with Docker
`--network host`, so the host must support that networking mode. CI runs on Linux;
local Docker Desktop results are recorded in `docs/verification.md`. The isolated
session-revocation fixture also uses the platform-specific addressing described above.

## Test design and source evolution

Use fixed identities and supplied timestamps in unit tests. Real infrastructure
proves transaction and delivery claims; in-memory doubles only prove application
policy. Bounded test observation is allowed; it is not permission to add browser
polling. Do not replace convergence assertions with fixed sleeps.

Every feature needs a registered owner/lane pairing and stable scenario IDs.
`pnpm check:specifications` rejects empty/malformed features, empty outlines,
duplicate IDs and unsupported execution filters. Undefined and pending steps
fail native runners. Add binding definitions in the owning service's language.
The report gate matches executed identities/example counts against the catalogue
and rejects missing, repeated, skipped or failed examples.

`pnpm generate:contracts` regenerates Go/Python bindings, TypeScript descriptors, OpenAPI bundles,
Python `.pyi` declarations, canonical fixture copies and migration checksum metadata. Review all generated
diffs. `pnpm check:large-file-review` checks the explicit generated/lockfile
exceptions; no handwritten exception currently exists.

Operations' type gate rejects missing command/event fields, incorrect field
types, invalid lifecycle literals and a port bound to another aggregate's
snapshot. Runtime tests separately exercise untyped HTTP bodies, corrupt stored
outcomes and restoration failures. A real RabbitMQ fixture checks that converting
Protobuf prices into typed integers preserves the original receipt material and
wire fingerprint. Generated declarations are compiler-owned; the two narrow
Pika call suppressions document the external stub's incorrect integer-only
timeout annotation and are checked for becoming unnecessary.

Executed local results and remaining limits are recorded in `docs/verification.md`.
Remote CI, cluster availability, load tests, actual email, deployment and independent
human UAT must be reported separately; this repository has no release lane.

`pnpm check:contracts` fails for changed, missing and newly generated bindings.
`pnpm verify` includes this gate and frontend OpenAPI mutation tests: renaming a
required request or response field must break real feature compilation, while
optional additions remain compatible. Required headers and query parameters must
break real callers, including shared command/list helpers. Protobuf generation
checks compile overlapping context-local tags/names and multiline field options. Native Go/Python/TypeScript checks
cover required annotations and explicit zero input values. PostgreSQL integration
checks cover command reply encoding rollback, recovered receipt outcomes and
immutable exact response bytes. Every owner language abandons and reclaims a
real reply lease before reply expiry; the Go lane also fences stale publishers.

Command-boundary regression checks use TypeScript compiler symbols, Go AST/type
information and Python AST capability/alias analysis. Negative fixtures exercise
mutation from event handlers and captured methods. DI graph tests resolve every
context provider or validate the Fx graph without infrastructure. Private command
codec tests preserve source receipt material with fixed values. The infrastructure
lane additionally verifies receiving-record rollback, immutable exact bytes, fenced
command dispatch recovery, bounded command retries and replay without a second
aggregate transition. Run `pnpm verify` and `pnpm test:integration` together for a
complete hand-off; static analysis assumes the supported typed coding conventions.

Command layout checks require one DTO/handler pair per file in each context's
`application/commands/` directory. Negative fixtures cover grouped pairs, a
missing handler and the previous flat layout. Go checks include nested
application packages; native scenario runners discover the relocated tests.

Omit the context from a focused command for service-wide native tests.
Full verification includes
historical OpenAPI/Protobuf comparison and negative fixtures for handler chaining,
repeated store calls, subscription completeness and lifecycle failure cleanup.
Infrastructure verification also stops a service to test readiness and inspects
a completed workflow without consuming its messages. See decision 0012.

Query layout checks require one input/handler pair under `application/queries/`.
Reaction checks require one event or projection handler per role module. Context
navigation checks compare the catalogue with source ownership and resolve README
links. Negative fixtures cover grouped queries, misplaced reactions, shared
adaptors importing business implementations and queries importing domain types.
Read-boundary tests preserve missing rows, revisions and continuations and reject
corrupt stored state. Python focused discovery recursively includes the owner's
`tests/contexts/<context>/` tree and its Gherkin binding.
