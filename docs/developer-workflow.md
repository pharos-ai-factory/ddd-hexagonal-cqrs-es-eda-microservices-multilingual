# Developing a use case

Start with `pnpm dev:doctor`. It reports installed tools, Docker access,
configuration availability and occupied ports without printing credentials.
Run `pnpm install --frozen-lockfile`, then `pnpm dev:up`. Linux remains the tested
infrastructure host; see TESTING.md for Docker Desktop limitations.

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
A command changes one root through one store call. Its domain method owns the
rule. A query reads through a query port. An event handler translates a fact into
a durable owner command; the receiving command handler changes the root later.

Every context puts commands in `application/commands/`, with one command DTO and
its matching handler in the same file. Name the file after the business action:
Go and Python use `create_order.go` / `accept_order.py`; TypeScript uses
`credit-collection.ts`. Import the concrete command package or module directly.
Shared errors and event DTOs have their own files. Event reactions and query
handlers remain separate from command execution. See decision 0013.

Create a starter, for example:

```sh
pnpm scaffold command ordering CancelOrder
pnpm scaffold query preparation FindTicket
pnpm scaffold subscription loyalty CreditPurchase
```

The result is under `.local/scaffolds/<context>/<name>/`. An existing output is
never overwritten. Review the inputs, replace generic snapshot types with the
owner's type, and place the code in the existing application layout. The generated
test deliberately fails until you specify observable behaviour. Registration
snippets identify the composition change; subscription starters additionally
include a manifest entry and codec/wiring checklist. They do not allocate wire
field numbers or silently register unfinished policy.

## Go: add an Ordering command

1. Add the named command DTO and handler together in
   `services/storefront/contexts/ordering/application/commands/cancel_order.go`.
   Use `package commands`; follow `change_quantity.go` for the transaction shape.
2. Implement and test the root behaviour in `domain/order.go`. Restore the root
   inside the store decision, call its method, return its snapshot and publications.
3. Register the handler in `apps/storefront/ordering.go`, receiving
   `AggregateCommandPort[domain.OrderState]` through the Fx provider.
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
   Use `AggregateCommandPort[TicketSnapshot]`; make one direct `execute` call.
2. Change `PreparationTicket` through its public behaviour and test the invariant.
3. Bind the handler as a provider in `apps/composition/preparation.py`.
4. For an event reaction, use a typed definition from `apps/incoming_events.py`.
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
   `apps/composition/loyalty.ts`. Keep concrete resources inside composition.
4. For a reaction, add the typed event name/payload association in `EventPayloads`,
   map it in the event handler, and add its private command codec/manifest/fixture.
5. Register through `commandSubscription` and validate the complete context set.
   The type checker rejects a handler for an incompatible event payload.
6. Resolve the provider in a composition test and run the focused Loyalty lane.

## Validate and follow the work

```sh
pnpm generate:contracts
pnpm check:compatibility                  # Accepted durable-command baseline
pnpm check:compatibility --against HEAD   # Compare a local change with this checkout
pnpm workflow:inspect <command-or-correlation-uuid>
pnpm events:replay loyalty.issue-reward.command
pnpm verify
pnpm test:integration
```

Inspection shows durable acceptance separately from publication and execution.
Dead-letter totals describe the queue, so a pending outcome may need further
investigation. Replay is a separate explicit operation and preserves identities.
The inspector caps each result set at 200 records and reports truncation.

CI compares compatibility with the PR base. Private queued-byte fixtures and native
codec tests protect stored work. Update the accepted decision, contributor guide
and executable example together whenever a convention changes.
