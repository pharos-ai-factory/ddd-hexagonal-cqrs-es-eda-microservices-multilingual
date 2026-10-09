# Developing a use case

Start with `pnpm dev:doctor`. It reports installed tools, Docker access,
configuration availability and occupied ports without printing credentials.
Run `pnpm install --frozen-lockfile`, then `pnpm dev:up`. CI runs on Linux;
see [TESTING.md](../TESTING.md) for host-networking requirements and
[executed verification](verification.md) for local results.

## Find the business capability

Run `pnpm context` for the context catalogue or `pnpm context ordering` to list
its current use cases, contracts, scenarios and focused test command. Each
context has a README at its source root. The catalogue and its links are checked
by `pnpm check:architecture`.

Read the [current rules](../CONTRIBUTING.md) and the affected entries in the
[decision index](decisions/README.md). Before editing, identify the immediate
invariant, aggregate owner, use-case role and any eventual downstream step.
Use a nearby implementation as the starting point:

| Role | Go | Python | TypeScript |
| --- | --- | --- | --- |
| Command and handler | [PlaceOrder](../services/storefront/contexts/ordering/application/commands/place_order.go) | [AcceptOrder](../services/operations/src/operations/contexts/preparation/application/commands/accept_order.py) | [CreditCollection](../services/engagement/src/contexts/loyalty/application/commands/credit-collection.ts) |
| Query and handler | [GetOrder](../services/storefront/contexts/ordering/application/queries/get_order.go) | [GetTicket](../services/operations/src/operations/contexts/preparation/application/queries/get_ticket.py) | [GetAccount](../services/engagement/src/contexts/loyalty/application/queries/get-account.ts) |
| Event/projection reaction | [MenuPublished projection](../services/storefront/contexts/ordering/application/projections/menu_published.go) | [OrderPlaced reaction](../services/operations/src/operations/contexts/preparation/application/event_handlers/order_placed.py) | [RewardEarned reaction](../services/engagement/src/contexts/loyalty/application/event-handlers/reward-earned.ts) |
| Read persistence mapping | [Order read repository](../services/storefront/contexts/ordering/adaptors/postgres/order_read_repository.go) | [Ticket read repository](../services/operations/src/operations/contexts/preparation/adaptors/persistence/ticket_read_repository.py) | [Account read repository](../services/engagement/src/contexts/loyalty/adaptors/persistence/account-read-repository.ts) |
| DI composition | [Ordering Fx module](../services/storefront/apps/storefront/ordering.go) | [Preparation providers](../services/operations/src/operations/apps/composition/preparation.py) | [Loyalty Awilix container](../services/engagement/src/apps/composition/loyalty.ts) |

## The edit/test loop

```sh
pnpm test:focused storefront --context ordering
pnpm test:focused operations --context preparation
pnpm test:focused engagement --context loyalty
pnpm test:focused api
pnpm test:focused web
pnpm test:unit                       # Native tests across all services
pnpm dev:watch operations engagement # Rebuild/restart only changed services
pnpm dev:status                      # Database/session, broker and realtime readiness
```

Compose Watch requires a version supporting `develop.watch` (Compose 2.22+).
The overlay rebuilds affected services and preserves the ordinary ingress and
secret configuration. This initial workflow favours the same runnable artefacts
as integration tests; Go compilation and static frontend export still take time.
Regenerate contracts before watching a schema change. Changing a dependency
manifest requires updating its lockfile before the next rebuild.

Open a native test and use the provided VS Code launch configurations. Go uses
the Go extension, Python uses the Operations virtual environment and debugpy,
and TypeScript uses Node with tsx. These profiles debug unit tests; infrastructure
tests continue through the isolated integration runner.

```sh
pnpm format path/to/changed-file.ts path/to/changed-file.py
pnpm format:check path/to/changed-file.go
```

Without paths, formatting selects tracked changes relative to HEAD. Add new files
explicitly. Go uses gofmt, Python uses pinned Ruff, and TypeScript/JSON/YAML/CSS/
Markdown use pinned Prettier. Generated folders are excluded. Existing files can
have older formatting; review formatting separately from business changes and
keep the 450-line responsibility limit.

## Choose the owning aggregate

Name the business invariant, the context and the aggregate before adding code.
A command loads and saves one root through its named write repository. Its domain method owns the
rule. A query reads through a query port. An event handler translates a fact into
a durable owner command; the receiving command handler changes the root later.
Application handlers never call another command handler. Register each handler
through the central command executor; it supplies the repository and owns
transaction and delivery bookkeeping. A projection reaction updates read data without changing a root.

Every context puts commands in `application/commands/`, with one command DTO and
its matching handler in the same file. Name the file after the business action:
Go and Python use `create_order.go` / `accept_order.py`; TypeScript uses
`credit-collection.ts`. Import the concrete command package or module directly.
Shared errors and event DTOs have their own files. Event reactions and query
handlers remain separate from command execution. Queries use `application/queries/`,
with one named query DTO and matching handler per file. Their named read repository ports
and read models live in `application/ports/` and `readmodels/` (Go), `read_models/`
(Python) or `read-models/` (TypeScript). Each event reaction has its own module in
`event_handlers/` (Python), `event-handlers/` (TypeScript) or `projections/` (Go's
current reactions). See decisions 0013 and 0014.

Create a starter, for example:

```sh
pnpm scaffold command ordering CancelOrder
pnpm scaffold query preparation FindTicket
pnpm scaffold subscription loyalty CreditPurchase
```

The result is under `.local/scaffolds/<context>/<name>/`. An existing output is
never overwritten. Review the inputs, replace generic snapshot types with the
owner's type. Query starters include an application view and read repository port; reuse
an existing one when suitable. Place the code in the existing application layout.
The generated test deliberately fails until you specify observable behaviour. Registration
snippets identify the composition change; subscription starters additionally
include a manifest entry and codec/wiring checklist. They do not allocate wire
field numbers or silently register unfinished policy.

## Go: add an Ordering command

1. Add the named command DTO and handler together in
   `services/storefront/contexts/ordering/application/commands/cancel_order.go`.
   Use `package commands`; follow `change_quantity.go` for the load, mutate and save flow.
2. Implement and test the root behaviour in `domain/order.go`. Use `OrderWriteRepository` to load the root, call its method and save it.
   The aggregate records facts; the owner publication mapper selects outgoing messages.
3. Register the handler in `apps/storefront/ordering.go`, receiving
   the invocation-scoped `OrderWriteRepository` through `command.Bind`.
   Register the returned executor with the owner RabbitMQ adaptor.
4. For an API operation, update Ordering's OpenAPI and Protobuf request sources,
   owner messaging translation and API mapping. Run contract generation.
5. Add a Gherkin example and native binding. Run the focused Ordering lane.

Go's existing event handlers update projections. A future aggregate-changing Go
subscription needs an owner durable command port, private codec/outbox and paired
consumers, following the Operations and Engagement examples. A generated event
handler is a starting point for that explicit implementation.

## Python: add a Preparation command or subscription

1. Add a TypedDict command and named handler together in Preparation's
   `application/commands/<action>.py` module.
   Inject `TicketWriteRepository`; load the root, invoke its behaviour and save it.
2. Change `PreparationTicket` through its public behaviour and test the invariant.
3. Bind its factory through `CommandExecutor` in `apps/composition/preparation.py`.
4. For an event reaction, use the context's typed definition in
   `adaptors/messaging/incoming_event.py`.
   Map its payload in the application event handler and enqueue the owner command.
5. Add the private Protobuf payload, codec, stable subscription declaration and
   historical fixture. Register its pair with `command_subscription`, then pass
   the complete context set through `complete_subscriptions`.
6. Add scenario/type/codec tests and run the focused Preparation lane.

The simple context registration supplies the typed incoming event, target,
event-handler method and command-handler method to the common runtime helper.
It receives no Pika connection or SQL transaction.

## TypeScript: add a Loyalty command or subscription

1. Add the command DTO and plain handler together under Loyalty's
   `application/commands/<action>.ts`. Follow `credit-collection.ts`; its
   constructor receives application ports.
2. Implement the aggregate rule and its domain test.
3. Extend `LoyaltyDependencies` and the explicit Awilix factory in
   `apps/composition/loyalty.ts`. Wrap its factory with `bindCommand`.
   Select concrete adaptors in composition; keep
   decoding and restoration in the owning adaptor modules.
4. For a reaction, add the typed event name/payload association in `EventPayloads`,
   map it in the event handler, and add its private command codec/manifest/fixture.
5. Register through `commandSubscription` and validate the complete context set.
   The type checker rejects a handler for an incompatible event payload.
6. Resolve the provider in a composition test and run the focused Loyalty lane.

## Wire an aggregate-changing subscription

The shared runtime installs the event consumer and private command consumer as
one pair. The context supplies the typed event, target identity, reaction, command
codec and command handler. Keep these responsibilities explicit:

1. Add the plain owner command/handler pair and its domain behaviour first.
2. Define the incoming event-to-payload association at the boundary. In Python,
   use the owner's typed `IncomingEvent` definition; in TypeScript, use
   `EventPayloads` and the supported event name.
3. Put the reaction in its own application module. Inject `DurableCommandPort`
   and map the fact to the command. Constructor injection also supplies test doubles.
4. Add the private Protobuf command and codec in the owner's messaging adaptor.
   Use stable typed subscription IDs and add the `subscriptions.json` entry.
   Preserve historical fixtures and allocate unused field numbers.
5. Bind the concrete outbox and handlers in the context container. Register the
   pair with the shared helper and validate the entire manifest.
6. Run `pnpm generate:contracts`, composition/codec tests and real hand-off tests.

For example, the existing Preparation composition binds one subscription with:

```python
command_subscription(
    container.codec(), definitions, ORDER_PLACED,
    lambda event: derived_id("ticket", event["orderId"]),
    container.accepted().handle, container.accept().execute,
)
```

This is an excerpt from
[Preparation composition](../services/operations/src/operations/apps/composition/preparation.py).
It returns both workers; the composition passes its complete worker set through
`complete_subscriptions`. TypeScript uses `commandSubscription` and
`completeSubscriptions` in the same roles. A projection-only subscription uses
its projection delivery path rather than creating an aggregate command.

Acceptance commits the event receipt, exact mapped command bytes and dispatch
intent. Failure before that commit remains the event consumer's responsibility.
After commit, command dispatch/execution owns recovery. Business rejections are
recorded outcomes; infrastructure failures remain retryable. Test duplicate
acceptance, rollback, lost confirms/ACKs and replay through the real infrastructure
lane. See [decision 0011](decisions/0011-durable-commands-before-aggregate-mutations.md).

## Add or change a query

1. Define its input and handler in the owning `application/queries/` file.
2. Select the required fields in an application read model and name the read repository
   capability. Keep aggregate snapshots inside command/persistence code.
3. Implement the read repository under the context's persistence adaptor, validating stored
   state before mapping it. Preserve revisions and pagination continuations.
4. Map transport arguments to the query in the context query adaptor. Bind the
   read repository and named handlers in composition, then register the transport operation.
5. Update the owner OpenAPI/Protobuf boundary if its public shape changes and run
   generation. Test missing resources, storage failures and relevant read behaviour.

Name the port `<Resource>ReadRepository` and its concrete Python/TypeScript class
`Postgres<Resource>ReadRepository`. Go uses `postgres.New<Resource>ReadRepository`.
Use `_read_repository` filenames in Go/Python and `-read-repository` in TypeScript.
Keep shared snapshot restoration in its own owner persistence module. Use a named `WriteRepository` for authoritative aggregate reads and saves.
The central command executor supplies a fresh repository and owns the local
transaction, receipt/outcome records and outgoing intent. [Decision 0015](decisions/0015-explicit-persistence-role-names.md)
records the distinction and examples.

Python focused tests recursively discover `tests/contexts/<context>/` and run the
context's Gherkin binding. Go and TypeScript discover tests beside their subjects.
Cross-context transport conformance and infrastructure checks remain service-wide.

## Change an HTTP operation or browser feature

1. Edit the owning OpenAPI fragment under `contracts/<context>/http_api/`.
   For an API-to-owner operation, also update the owner's published Protobuf
   request/reply sources. These describe separate boundaries.
2. Update the API's explicit translation and owner messaging adaptors.
   Keep generated messages outside application/domain packages. The API sends
   business commands and queries through RabbitMQ. Owner services expose only
   health and diagnostics over HTTP; inspect business state through the API.
3. Run `pnpm generate:contracts` for a combined change, or `pnpm generate:http`
   for OpenAPI alone. Review generated frontend types, Go bundles and wire mappings.
4. Update the feature and relevant conformance examples. Feature requests use
   `adaptors/http/client`; incompatible schema changes must fail real callers.
5. Run generation/compatibility checks, the focused service/web tests and both
   hand-off gates. The integration lane exercises the actual browser and owner DTOs.

Browser task folders are `menu`, `ordering`, `preparation`, `collection`, `rewards`
and `notifications` under `services/web/src/features/`. The page composes them.
Shared command recovery, realtime reconciliation and display helpers live under
`shared/`. Preserve session-authorised channels, per-root revision guards and
the saved uncertain command across reload/sign-in. Reconciliation runs after
subscription attachment or a history gap. Business polling and client-selected
channels remain prohibited. See [the browser contract](../CLIENT-SUBSCRIPTIONS.md).

## Keep the DI boundary explicit

Fx, Dependency Injector and Awilix belong in `apps/`. Resolve the context graph
before starting workers and retain plain constructor-injected ports inside
handlers. A handler or domain object never resolves a container. New resources
need disposal and partial-startup failure coverage; shutdown drains workers before
closing pools. Extend the existing composition graph tests when adding providers.
The [DI decision](decisions/0010-dependency-injection-frameworks.md) records the
framework comparison and why these choices fit the reference.

## Validate and follow the work

```sh
pnpm check:contracts                     # Regenerate in isolation; fail on output drift
pnpm check:compatibility                  # Accepted durable-command baseline
pnpm check:compatibility --against HEAD   # Compare the working tree with HEAD
pnpm workflow:inspect <command-or-correlation-uuid>
pnpm events:replay loyalty.issue-reward.command
pnpm verify
pnpm test:integration
```

Generation and compatibility answer different questions: fresh generated files
can still represent an incompatible interface. The [testing guide](../TESTING.md)
maps each rule to its check and distinguishes fast policy tests from live evidence.

Inspection shows durable acceptance separately from publication and execution.
Dead-letter totals describe the queue, so a pending outcome may need further
investigation. Replay is a separate explicit operation and preserves identities.
The inspector caps each result set at 200 records and reports truncation.

CI compares compatibility with the PR base. Private queued-byte fixtures and native
codec tests protect stored work. Before hand-off, review `git diff --check`, the
generated diff, new navigation links and any changed architectural assumptions.
Update the accepted decision, agent/contributor guidance and executable examples
together whenever a convention changes. Record actual checks and limitations in
`docs/verification.md`, then describe the concrete behaviour and evidence in the PR.
