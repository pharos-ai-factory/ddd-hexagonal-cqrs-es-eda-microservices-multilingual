# Verification

`pnpm verify` is the deterministic gate for every application:

- Service/context ownership and inward imports across Go, Python and TypeScript;
  policy negative fixtures and large-file responsibility checks.
- Formatting, race-enabled unit tests and vet for both Go modules.
- Frozen Python dependencies, Ruff, strict Mypy over handwritten source/tests,
  type-gate negative examples and domain/contract tests.
- Engagement and web TypeScript checks and behaviour tests.
- 51 executable Gherkin scenarios through native Go/Python/TypeScript handlers;
  catalogue validation, typed step bindings and a dry run of workflow bindings.

`pnpm test:bdd` runs only the fast Gherkin scenarios. The
[specification catalogue](specifications/README.md) explains tags, ownership,
individual scenario selection and reports under `.local/bdd/`.
The dry run binds workflow steps without executing them and is not evidence
that the live workflows passed.

`pnpm test:integration` provisions a disposable Compose project with random ports
and separate credentials. It builds all applications, including the Next.js
production build used locally, and runs:

| Boundary | Required evidence |
| --- | --- |
| Go PostgreSQL | State/receipt/outbox atomicity, terminated connection, concurrent commands, conflicting consumer identities, restricted roles, immutable evidence and fenced leases |
| Python and TypeScript PostgreSQL | Encoder-failure rollback, recorded retry outcomes, conflicting input, expected-version races, atomic browser publications and realtime lease fencing |
| RabbitMQ | Protobuf metadata, confirmed mandatory publication, commit-before-ACK redelivery, bounded retry, dead-letter preservation, replay and private-topic restrictions |
| Cross-language workflow | Menu → order → preparation → collection → account → reward → notifications; Operations outage and paused private Reward consumer |
| Gherkin PostgreSQL | Recorded `menu_pending` after projection arrival, new-attempt success and conflicting key reuse; atomic browser intent |
| Gherkin live workflows | Ten focused scenarios: repeated business facts with new event IDs, duplicate grants/notifications, concurrent collections/redemptions, customer isolation, lost provider response and recorded completion rejection |
| Provider | Acceptance followed by a lost response creates one provider effect |
| Browser / Centrifugo | Two independent windows, persisted uncertain command retry across authentication loss, typed live updates, no business polling, recovered history, missing-history reconciliation, mobile layout and logout |
| Valkey authority | Realtime credentials cannot write session keys or administer ACL users; permission probes require a disposable project |

Fixtures use real PostgreSQL and RabbitMQ. Missing infrastructure is a failure,
not a skipped pass. Component fixtures are cleared only in the guarded disposable
project before the acceptance journey; the development dataset is preserved.
The test project is removed at completion, including on failure. Use
`python3 scripts/integration.py --keep` to retain it for diagnosis.

Gherkin duplicate scenarios wait for the new message's committed consumer receipt
before asserting no second effect. Cross-service fixtures use only the Go API,
published wire contracts and explicit database evidence; they do not import
another service's domain code. Fresh identities isolate scenarios. Their shared
provider failure control requires serial execution in one disposable project.

`pnpm test:browser` runs the browser scenario against the existing development
stack. It adds demonstration data and removes test channel history during the
recovery exercise. `CAFE_ENV_FILE` selects a different repository-local env file.
Chromium runs in the pinned Playwright Docker image, including its system
libraries; the wrapper handles rootless Docker ownership. Linux host networking
is currently required for this lane.

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

`pnpm generate:contracts` regenerates Go/Python bindings, TypeScript descriptors,
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
