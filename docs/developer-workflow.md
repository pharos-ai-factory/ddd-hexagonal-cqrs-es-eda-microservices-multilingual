# Developing a use case

Start with `pnpm dev:doctor`. It reports installed tools, Docker access,
configuration availability and occupied ports without printing credentials.
Run `pnpm install --frozen-lockfile`, then `pnpm dev:up`. CI runs on Linux;
see [TESTING.md](../TESTING.md) for host-networking requirements and
[executed verification](verification.md) for local results.

## Find the business capability

`pnpm docs:dev` serves the repository guides at `http://127.0.0.1:3001` without
starting the café services. Run `pnpm docs:check` after documentation changes;
the complete static gate also runs it. The [documentation guide](documentation.md)
explains navigation, diagrams, source links and generated Markdown output.

Run `pnpm context` for the context catalogue or `pnpm context ordering` to list
its current use cases, contracts, scenarios and focused test command. Each
context has a README at its source root. The catalogue and its links are checked
by `pnpm check:architecture`.

Read the [current rules](../CONTRIBUTING.md) and the affected entries in the
[decision index](decisions/README.md). Before editing, identify the immediate
invariant, aggregate owner, use-case role and any eventual downstream step.
Use a nearby implementation as the starting point:

| Role                      | Go                                                                                                                       | Python                                                                                                                              | TypeScript                                                                                                             |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| Command and handler       | [PlaceOrder](../services/storefront/contexts/ordering/application/commands/place_order.go)                               | [AcceptOrder](../services/operations/src/operations/contexts/preparation/application/commands/accept_order.py)                      | [CreditCollection](../services/engagement/src/contexts/loyalty/application/commands/credit-collection.ts)              |
| Query and handler         | [GetOrder](../services/storefront/contexts/ordering/application/queries/get_order.go)                                    | [GetTicket](../services/operations/src/operations/contexts/preparation/application/queries/get_ticket.py)                           | [GetAccount](../services/engagement/src/contexts/loyalty/application/queries/get-account.ts)                           |
| Event/projection reaction | [MenuPublished projection](../services/storefront/contexts/ordering/application/projections/menu_published.go)           | [OrderPlaced reaction](../services/operations/src/operations/contexts/preparation/application/event_handlers/order_placed.py)       | [RewardEarned reaction](../services/engagement/src/contexts/loyalty/application/event-handlers/reward-earned.ts)       |
| Read persistence mapping  | [Order read repository](../services/storefront/contexts/ordering/adaptors/persistence/postgres/order_read_repository.go) | [Ticket read repository](../services/operations/src/operations/contexts/preparation/adaptors/persistence/ticket_read_repository.py) | [Account read repository](../services/engagement/src/contexts/loyalty/adaptors/persistence/account-read-repository.ts) |
| DI composition            | [Ordering Fx module](../services/storefront/apps/storefront/di/ordering.go)                                              | [Preparation providers](../services/operations/src/operations/apps/composition/preparation.py)                                      | [Loyalty Awilix container](../services/engagement/src/apps/composition/loyalty.ts)                                     |

## The edit/test loop

```sh
pnpm check                           # All static checks; no live services
pnpm check operations                # Complete owning project and boundaries
pnpm check --changed                 # Edited owners, including new/deleted files
pnpm check:watch                      # Initial checks, then checks after edits
pnpm test:focused storefront --context ordering
pnpm test:focused operations --context preparation
pnpm test:focused engagement --context loyalty
pnpm test:focused api
pnpm test:focused web
pnpm test:unit                       # Native tests across all services
pnpm dev:watch operations engagement # Rebuild/restart only changed services
pnpm dev:status                      # Database/session, broker and realtime readiness
```

Focused tests first run the owning service's static checks. Shared contracts or
tooling edits select every service with `pnpm check --changed`. The watcher
debounces changes and continues after failures, including edits made during a run.
`dev:watch` runs the same watcher beside Compose and stops it on exit.

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

Without paths, formatting selects tracked changes and new untracked files.
Use `pnpm format --all` to apply the baseline or `pnpm format:check --all` to
check every handwritten file, including on a clean checkout. `--service web`
selects one complete service. Go uses gofmt, Python uses pinned Ruff, and
TypeScript/JSON/YAML/CSS/Markdown use pinned Prettier. Generators own generated
outputs. Generated filenames include `generated`, or live under a `generated`
folder; dependency lockfiles and Next.js' required `next-env.d.ts` retain their
standard names. Keep the 450-line responsibility limit when formatting source.

VS Code recommends the corresponding extensions and enables formatting on save,
Mypy/ESLint diagnostics and gopls analysis. The task list includes changed-project
checks and the continuous watcher. CLI checks are authoritative for every editor.
See [decision 0018](decisions/0018-static-checks-and-formatting.md).

## Choose the owning aggregate

Name the business invariant, the context and the aggregate before adding code.
A command loads and saves one root through its named write repository. Its domain method owns the
rule. A query reads through a query port. An event handler translates a fact into
a durable owner command; the receiving command handler changes the root later.
Application handlers never call another command handler. Register each handler
through the shared command registration API, naming the command and plain handler
factory together. Its internal executor supplies the repository and owns transaction
and delivery bookkeeping. A projection reaction updates read data without changing a root.

Every context puts command handlers in `application/commands/`. Each imports its
plain command payload from the schema-generated service-local request contracts. Name the file after the business action:
Go and Python use `create_order.go` / `accept_order.py`; TypeScript uses
`credit-collection.ts`. Import the concrete command package or module directly.
Shared errors and event DTOs have their own files. Event reactions and query
handlers remain separate from command execution. Queries use `application/queries/`,
with one handler per file importing its schema-generated query input. Their named read repository ports
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
owner's type. Query starters include a response schema, generated view alias and read repository port; reuse
an existing one when suitable. Place the code in the existing application layout.
The generated test deliberately fails until you specify observable behaviour. Registration
snippets identify the composition change; subscription starters additionally
include a manifest entry and codec/wiring checklist. They do not allocate wire
field numbers or silently register unfinished policy.

## Go: add an Ordering command

The [command recipe](education/concepts/commands.md) shows the exact schema,
generated identifier, parser and registration for each language.

1. Define the command payload in `contracts/ordering/messaging/commands/v1/`,
   including `owner: "ordering"` and `visibility: PUBLIC`, and add its case in
   `Command.payload` under `contracts/ordering/messaging/v1/`.
2. Run `pnpm generate:contracts`. Inspect the plain input under
   `services/storefront/contracts/requests/generated/model/ordering/`, the wire bindings,
   typed identifier and generated owner input mapping.
3. Import that generated input into the handler in
   `services/storefront/contexts/ordering/application/commands/cancel_order.go`.
   Use `package commands`; follow `change_quantity.go` for load, mutate and save.
4. Implement and test the root behaviour in `domain/order.go`. Use `OrderWriteRepository`
   to load the root, call its method and save it. Keep additional semantic input
   validation in a handwritten owner wrapper around the generated mapping.
5. Register the generated command name, input mapping and handler factory together in
   `apps/storefront/di/ordering_requests.go` with `registration.Register`. The
   aggregate registry supplies its transaction and invocation-scoped repository.
   Use `RegisterProjection` when immutable projection data must be prepared first.
6. For an API operation, update Ordering's OpenAPI and API field mapping sources,
   then rerun contract generation.
7. Add a composition dispatch test, Gherkin example and native binding. Run the focused Ordering lane.

For an event-triggered owner command, follow the [RecordCollection recipe](education/concepts/durable-commands.md#go-record-collection-on-the-order). Define the private
Protobuf command and codec, declare the source subscription, bind `CommandBusPort`
to `NewCommandOutbox`, and register both consumers with `registration.Subscribe`.
Generation supplies typed identifiers, topology and broker permissions. The
command handler receives its scoped repository and realtime bus through the same
aggregate registry used by published commands.

## Shared boundary infrastructure

Start with [the service and context adaptor map](adaptor-layout.md).
Shared `persistence/` modules own database access, exact outbox/reply bytes and
lease updates. `messaging/` modules own envelopes and RabbitMQ delivery;
`realtime/` owns browser encoding and gateway publication; `http/` owns operational
security. Context-specific parsing and restoration stay in the context.

Request envelope factories come from the declared request families. Update their
generator and run `pnpm generate:contracts` when extending that catalogue.
[Decision 0021](decisions/0021-service-and-context-adaptor-layout.md) records the layout.

## Python: add a Preparation command or subscription

1. Define the command payload and envelope case in Preparation's Protobuf schema,
   declaring its owner and visibility. Generate the plain TypedDict, then import it
   into the named handler in `application/commands/<action>.py`.
   Inject `TicketWriteRepository`; load the root, invoke its behaviour and save it.
2. Change `PreparationTicket` through its public behaviour and test the invariant.
3. For a published command, add its payload and `Command.payload` case under
   `contracts/preparation/messaging/`, then run `pnpm generate:contracts`.
   Use its generated input mapper in `adaptors/messaging/generated/request_payloads.py`.
   Add owner semantic validation in `ticket_requests.py` when needed. See the [complete recipe](education/concepts/commands.md).
4. In `apps/composition/preparation.py`, use the configured aggregate registration:
   `commands.request(registry, PreparationCommand.START_PREPARATION, wire.start, StartPreparationCommandHandler)`.
   The shared helper creates the executor and supplies each invocation's repository.
5. For a private command triggered by an event, follow the subscription path
   instead of the published-request steps. Use the context's typed definition in
   `adaptors/messaging/order_placed_subscription.py`.
   Map its payload in the application event handler and enqueue the owner command.
6. Add the private Protobuf payload, codec, stable subscription declaration and
   historical fixture. Register its codec and handler factory through `container.commands().subscription`, then pass
   the complete context set through `complete_subscriptions`.
7. Run contract generation for private schema changes. Add composition dispatch,
   scenario/type/codec tests and run the focused Preparation lane.

The simple context registration supplies the typed incoming event, target,
event-handler method and plain command-handler factory to the common registration helper.
It receives no Pika connection or SQL transaction.

## TypeScript: add a Loyalty command or subscription

1. Define the command payload and envelope case in Loyalty's Protobuf schema,
   declaring its owner and visibility. Generate the plain type, then import it
   into the handler under `application/commands/<action>.ts`. Follow `credit-collection.ts`; its
   constructor receives application ports.
2. Implement the aggregate rule and its domain test.
3. For a published command, add its payload and `Command.payload` case under
   `contracts/loyalty/messaging/`, then run `pnpm generate:contracts`.
   Use its generated parser in `adaptors/messaging/generated/request-payloads.ts`.
   Retain owner semantic validation in `reward-requests.ts`. See the [complete recipe](education/concepts/commands.md).
4. Extend `LoyaltyDependencies` and the explicit Awilix factory in
   `apps/composition/loyalty.ts` for any new dependency. Register the command name,
   parser and handler factory together through `c.rewardCommands.request(...)`
   or the aggregate registration's `subscription(...)` method.
   Select concrete adaptors in composition; keep
   decoding and restoration in the owning adaptor modules.
5. For a private command triggered by an event, follow the subscription path
   instead of the published-request steps. Use the generated published name/payload association,
   map it in the event handler, and add its private command codec/manifest/fixture.
6. Register the codec and plain handler factory through the aggregate registration's
   `subscription` method and validate the complete context set.
   The type checker rejects a handler for an incompatible event payload.
7. Run contract generation for private schema changes. Resolve the provider and
   dispatch the command in a composition test, then run the focused Loyalty lane.

## Wire an aggregate-changing subscription

The shared runtime installs the event consumer and private command consumer as
one pair. The context supplies the typed event, target identity, reaction, command
codec and command handler. Keep these responsibilities explicit:

1. Add the plain owner command/handler pair and its domain behaviour first.
2. Define the incoming event-to-payload association at the boundary. In Python,
   use the owner's typed `IncomingEvent` definition; in TypeScript, use
   `EventPayloads` and the supported event name.
3. Put the reaction in its own application module. Inject `CommandBusPort`
   and map the fact to the command. Constructor injection also supplies test doubles.
4. Add the private Protobuf command and codec in the owner's messaging adaptor.
   Use stable typed subscription IDs and add the `subscriptions.json` entry.
   Preserve historical fixtures and allocate unused field numbers.
5. Bind the concrete outbox and handlers in the context container. Register the
   pair with the shared helper and validate the entire manifest.
6. Run `pnpm generate:contracts`, composition/codec tests and real hand-off tests.

For example, the existing Preparation composition binds one subscription with:

```python
container.commands().subscription(
    container.codec(), definitions, ORDER_PLACED,
    lambda event: derived_id("ticket", event["orderId"]),
    container.accepted().handle, AcceptOrderCommandHandler,
)
```

This is an excerpt from
[Preparation composition](../services/operations/src/operations/apps/composition/preparation.py).
It returns both workers; the composition passes its complete worker set through
`complete_subscriptions`. TypeScript uses `c.accountCommands.subscription` and
`completeSubscriptions` in the same roles. Both helpers construct executors internally. A projection-only subscription uses
its projection delivery path rather than creating an aggregate command.

Acceptance commits the event receipt, exact mapped command bytes and dispatch
intent. Failure before that commit remains the event consumer's responsibility.
After commit, command dispatch/execution owns recovery. Business rejections are
recorded outcomes; infrastructure failures remain retryable. Test duplicate
acceptance, rollback, lost confirms/ACKs and replay through the real infrastructure
lane. See [decision 0011](decisions/0011-durable-commands-before-aggregate-mutations.md).

The command name or private codec and its concrete handler factory appear in the
same registration. Dependencies are resolved at startup. For provider evidence,
Communication uses `preparedSubscription` and `NotificationDeliveryPreparation`
to obtain the receipt before starting the aggregate transaction. See
[decision 0023](decisions/0023-direct-command-registration.md).

## Add or change a query

1. Define the input and response in owner Protobuf. Set `contract.response` to the
   fully qualified success message, then add both envelope alternatives.
2. Run `pnpm generate:contracts`. Use the generated resource DTO at the named read
   repository port; existing read-model modules can alias it.
3. Validate stored state and select published fields in the owner persistence adaptor.
4. Return the generated query response from the handler. Select `not_found` explicitly,
   preserve revisions and retain complete/paginated read behaviour.
5. Bind the concrete reader and handler in DI, then call the generated owner registration
   with the parser and handler. It fixes the identifier, response type and encoder.
6. Update the HTTP boundary, generate and test successful/missing reads, repository
   failures, presence, nested fields and pagination through the registered route.

Preparation's actual binding is:

```python
responses.register_get_ticket(registry, wire.get_ticket, container.get_ticket().execute)
```

`responses` imports the owner's generated `response_payloads` module. Go uses
`responses.RegisterGetOrder`; TypeScript uses `responses.registerGetReward`. Follow the
[complete three-language recipe](education/concepts/queries.md) and
[response recipe](education/concepts/responses.md). The scaffold supplies request
and reply schema fragments and imports their generated types.

## Change an HTTP operation or browser feature

1. Edit the owning OpenAPI fragment under `contracts/<context>/http_api/`.
   For an API-to-owner operation, also update the owner's published Protobuf
   request/reply sources. These describe separate boundaries.
2. Update the API's explicit translation and owner messaging adaptors.
   Keep generated messages outside application/domain packages. The API sends
   business commands and queries through RabbitMQ. Owner services expose only
   health and diagnostics over HTTP; inspect business state through the API.
3. Give each operation one declared tag: its context name for business HTTP, or
   `session`, `realtime` or `operational` for technical HTTP. Run
   `pnpm generate:contracts` for a combined change, or `pnpm generate:http` for
   OpenAPI alone. Review the Orval tag modules, models, MSW/Faker outputs, Go
   bundles and wire mappings.
4. Update the feature and relevant conformance examples. Queries use the generated
   TanStack functions, hooks or query options. Durable commands use
   `adaptors/http/operation-client`, which retains their raw response while taking
   its operation and input/output types from the generated manifest. Incompatible
   schema changes must fail real callers.
5. Run generation/compatibility checks, the focused service/web tests and both
   hand-off gates. The integration lane exercises the actual browser and owner DTOs.

Browser task folders are `menu`, `ordering`, `preparation`, `collection`, `rewards`
and `notifications` under `services/web/src/features/`. The page composes them.
Shared command recovery, realtime reconciliation and display helpers live under
`shared/`. Preserve session-authorised channels, per-root revision guards and
the saved uncertain command across reload/sign-in. Reconciliation runs after
subscription attachment or a history gap. Business polling and client-selected
channels remain prohibited. See [the browser contract](../CLIENT-SUBSCRIPTIONS.md).
See the [browser OpenAPI client guide](openapi-client.md) for generated paths,
cache ownership and MSW/Faker examples.

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

Generation and its native compiler regression tests install the pinned Go
Protobuf plugin automatically when `.tools/bin/protoc-gen-go` is absent or has
the wrong version. The first run requires access to the pinned Go module; a
previous local generation run is unnecessary. The shared setup lives in
`scripts/protobuf_tools.py` and verifies the installed executable before use.

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

## Add a published integration event

Define the event once in its producer's Protobuf schema with `cafe.events.event`
metadata and `cafe.events.rules` constraints. New schema files in an existing
context family are discovered automatically. Run `pnpm generate:contracts` to
produce the plain DTO, typed conversion, publication factory and registry.
Queue a plain `*DomainEvent` on the aggregate and define its owner-private codec.
The repository adaptor stages that private event when saving. Add an independent
`*DomainEventHandler` that maps the delivered fact to the generated
`*IntegrationEvent` DTO through `IntegrationEventBusPort`. Its concrete outbox
adaptor uses the typed publication factory and commits the subscription receipt
with outgoing bytes and dispatch. ACK follows that commit; a relay handles the
later broker publication. Declare the private consumer's `publication` in its
manifest and bind it through the shared publication-subscription helper.

Declare each consumer in its own `adaptors/messaging/subscriptions.json` and bind
its handler explicitly. Add a service contract-family dependency when this is its
first use of another producer. The shared envelope and handwritten codecs require
no event-specific edits. Keep semantic evidence checks in context adaptors.
See [the event recipe](../contracts/shared/messaging/events.md) and
[decision 0017](decisions/0017-generated-event-boundaries.md).

## Changing broker subscriptions and authority

After editing an owner subscription or API request contract selection, run
`pnpm generate:contracts` and review the generated topology and
`services/storefront/apps/topology/permissions.generated.json`. Run `pnpm dev:up`
to apply the catalogue permissions with the dedicated topology credential. The
launcher restarts managed broker clients; restart separately launched host workers
to replace cached authorisation. See [broker permissions](broker-permissions.md)
for queue families, private-event ownership and real-broker verification.

## Publish browser messages explicitly

Follow the [end-to-end recipe](education/end-to-end-recipe.md) and
[realtime recipe](education/concepts/realtime-events.md). Inject `RealtimeBusPort`
into a command use case for atomic state/outcome publication, or declare a dedicated
`realtime` subscriber for a domain/integration event. Bind the bus and handler in
owner composition, generate contracts/topology, and validate the complete worker set.
A saved root produces no browser message unless application code selects it.
Event-driven refreshes publish the owner view's actual revision. Expected business
rejections may publish a matching typed outcome; technical failures remain retryable.
See [decision 0026](decisions/0026-explicit-bus-ports-and-realtime-publication.md).

Command and query payloads are schema-generated plain contracts. Declare owner and
visibility in Protobuf; generate types and mechanical mappings, then implement
handlers and semantic validation. See [decision 0028](decisions/0028-schema-generated-command-and-query-inputs.md).

## Generated package ownership

Place wholly generated packages beneath a `generated` directory and give files
ordinary descriptive names. Go request DTOs use
`contracts/requests/generated/model/<owner>/{payloads,responses}.go`; event DTOs
use `contracts/events/generated/model/<producer>.go`. Shared response encoders
and owner query registrations use adaptor `generated` packages. Mixed packages
retain `_generated.go` filenames and all generated source retains its header.

Update generator destinations, cleanup scopes, formatter globs, architecture
classification, imports, scaffolds and source-linked documentation together.
Request wire cleanup owns only `generated/cafe/requests`, so it preserves plain
models. Run `pnpm generate:contracts` and `pnpm check:contracts` to verify the
result. See [decision 0030](decisions/0030-generated-package-layout.md).
