# Executable behaviour specifications

The feature files are the language-neutral description of the café's behaviour.
Step definitions call each service's actual application handlers, or exercise the
running system through its Go API and published contracts.

| Specification | Runner | Boundary | Expanded scenarios |
| --- | --- | --- | --- |
| [Menu](menu/menu.feature) | Godog | Go application decisions | 12 |
| [Ordering](ordering/ordering.feature) | Godog | Go application decisions | 11 |
| [Projection recovery](ordering/projection-recovery.feature) | Godog | Go handlers and real PostgreSQL | 2 |
| [Preparation](preparation/preparation.feature) | pytest-bdd | Python application decisions | 6 |
| [Collection](collection/collection.feature) | pytest-bdd | Python application decisions | 5 |
| [Loyalty](loyalty/loyalty.feature) | Cucumber | TypeScript application decisions | 11 |
| [Communication](communication/communication.feature) | Cucumber | TypeScript application decisions | 6 |
| [Durable workflows](workflows/durable-workflows.feature) | Cucumber | Go API, three services, PostgreSQL, RabbitMQ and provider | 10 |

There are **63 executable scenarios**, counting each Scenario Outline example:
51 fast, two PostgreSQL and ten cross-service. `pnpm check:specifications`
parses the actual Gherkin and prints current counts.

## Run the specifications

From the repository root:

```sh
pnpm install --frozen-lockfile
pnpm test:bdd              # All 51 fast scenarios; no infrastructure required
pnpm verify                # Fast scenarios plus existing deterministic gates
pnpm test:integration      # Two PostgreSQL and ten workflow scenarios, plus
                           # existing fault-injection and real browser evidence
```

`pnpm test:integration` owns a disposable Compose project. It stops the domain
workers for component fixtures, clears those fixtures, restarts the workers and
runs the live workflows. The Gherkin harness refuses a development project.
No scenario truncates development data. Do not run its live suite concurrently
against the same test project: the provider's lost-response control is shared.

To focus on one rule:

```sh
python3 scripts/go.py test ./contexts/menu/application -run 'TestMenuFeatures/An_empty_edition'
uv run --project services/operations pytest services/operations/tests/bdd -k wrong_code
pnpm --filter @cafe/engagement test:bdd --tags @LOYALTY_006
```

The root runners write local Cucumber JSON under `.local/bdd/`:
`menu.json`, `ordering.json`, `operations.json`, `engagement.json`,
`ordering-postgres.json` and `workflows.json`. The deterministic gate's
`workflows-dry-run.json` checks step binding only; it is not integration evidence.
Reports are ignored by Git and are not sent to a hosted Cucumber service.

## What each boundary proves

Fast scenarios use fresh command/projection probes and supplied identities/time.
They prove the actual handler's decision, aggregate state, event mapping, rejection
or no-op. The probes intentionally do not implement receipt storage or locking.
They cannot establish atomicity, durable delivery or concurrency safety.

For example, the Loyalty scenarios observe a decision to earn a grant
without invoking Reward issuance. A separate step invokes IssueReward and proves
that it does not rewrite the account. Private/public visibility is selected by
the canonical event catalogue, and Go additionally asserts publication visibility.
The existing codec and architecture checks verify catalogue agreement.

The PostgreSQL scenarios deliberately withhold Ordering's menu projection. They
prove that `menu_pending` remains the stored outcome on identical retry, even
after the projection arrives. A new command can then create the order, with one
atomic browser publication. Conflicting reuse of a key cannot rewrite its owner.

The workflow scenarios replay an original committed Protobuf fact with a **new
event ID**, using the publishing context's restricted broker credentials. They
wait for the receiving consumer's committed receipt before asserting unchanged
state or absence of another effect. An empty queue or an arbitrary sleep would
not establish that the duplicate had actually been handled.

Concurrent collection commands exercise convergence from independent pickups
into one account. Concurrent reward redemptions exercise a real expected-version
race through the API. The existing adaptor tests separately race database writes,
terminate connections, inject encoder failures and fence expired dispatch leases.

The browser recovery scenario remains Playwright, and the service-outage/private
consumer-pause exercise remains the existing scripted journey. Both still run in
the integration gate; wrapping a whole journey in one Gherkin step would hide its
business behaviour.

## Add or change a rule

1. Describe one business decision or workflow outcome in its owner's feature.
   Use the vocabulary in [DDD.md](../DDD.md), with observable Given/When/Then
   steps. Put repeated data combinations in a Scenario Outline.
2. Give the Feature its owner and lane tags. Give each Scenario or Outline one
   stable identity, such as `@ORDER_012`. Keep IDs when rewording a scenario.
3. Implement steps in the owner's language. Go bindings live in
   `services/storefront/contexts/<owner>/application/*bdd*test.go`; Python bindings
   in `services/operations/tests/bdd`; TypeScript bindings in
   `services/engagement/tests/bdd`. Cross-service bindings live in `tests/acceptance`
   and never import a service's business implementation.
4. Choose the smallest boundary that proves the claim. A new infrastructure lane
   needs an explicit runner entry in the catalogue policy and verification script.
   Existing lanes are `@fast`, Ordering's `@postgres`, and `@integration` workflows.
5. Run the relevant scenario, then the required gates. Treat undefined or pending
   steps as failures. Do not add skip tags to make a feature pass.

The catalogue gate rejects malformed or empty features, empty outlines, duplicate
IDs, unsupported owner/lane combinations and extra execution-filter tags.
Runner strict modes reject undefined and pending steps.
After execution, the report check requires every registered scenario identity
and outline example to have passed. Missing, repeated or skipped examples fail
the gate.
