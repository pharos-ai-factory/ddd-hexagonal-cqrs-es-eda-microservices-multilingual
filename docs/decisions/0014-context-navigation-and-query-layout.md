# 0014 — Context navigation and explicit read use cases

Status: accepted, 9 October 2026. Extends decisions 0012 and 0013.

## Decision

Keep bounded contexts within their owning service. Within a context, name files
for business responsibilities and group them by their application role:

```text
contexts/ordering/
  README.md
  domain/order.go
  application/
    commands/create_order.go
    commands/place_order.go
    queries/get_order.go
    queries/list_orders.go
    readmodels/order.go
    ports/order_reader.go
    projections/menu_published.go
  adaptors/
    http/
    messaging/
    postgres/order_reader.go
    queries/order.go
```

Python uses `read_models` and `event_handlers`; TypeScript uses `read-models`
and `event-handlers`. Go uses ordinary package names such as `readmodels` and
`projections`. Each query file contains its named input and matching handler.
Each event reaction or projection handler has its own business-named file.
Python domain modules name their aggregate and package initialisers contain
only documentation. Command layout follows decision 0013.

Application queries depend on named reader ports and application-owned read
models. Persistence adaptors validate stored authority before selecting view
fields. Shared version/page containers carry those values without imposing a
particular storage format. The current small reference reads current-state
snapshots; dedicated read tables can implement the same reader ports later.

Context query adaptors translate HTTP/RabbitMQ arguments into named queries.
The generic transport helpers still handle errors, pagination and envelopes.
Context-owned persistence modules restore the aggregate and map its view.
Context-owned messaging modules parse private command payloads and own stable
subscription definitions. Composition binds these capabilities and their
resource lifetimes through the existing DI frameworks.

Shared technical adaptors can consume generated schema catalogues. Handwritten
context decoding, restoration and business mappings belong under that context.
Query handlers and read models do not import aggregate implementations.

## Developer navigation

`docs/contexts.json` records source ownership and browser features.
`pnpm context` lists the six contexts; `pnpm context <name>` discovers their
current use-case files and prints the focused test command. Each context README
links domain behaviour, commands, queries, ports, reactions, adaptors, composition,
contracts, specifications and tests. Checks reject stale ownership or broken
navigation targets.

Go and TypeScript tests sit beside their subjects. Python tests mirror context
ownership under `tests/contexts/<context>/`; focused discovery selects that tree
and the context's Gherkin bindings. Service-wide transport conformance and
infrastructure tests remain at service level. Scaffolds follow these role folders
and include the query input, reader port, view and an intentionally failing
behaviour test.

The browser groups features by user task: menu, ordering, preparation, collection,
rewards and notifications. Its page composes the journey. Shared command recovery,
realtime reconciliation and display helpers remain under `shared/`; generated
contracts and HTTP access remain under `adaptors/`. The aggregate-to-projection
catalogue remains shared because one authorised subscription session serves the
complete café workspace.

## Consequences

Developers can start from a business capability, find one use case and follow
its dependencies to the boundary. Read models can evolve independently of
aggregate storage. This adds explicit mappings, small modules and DI bindings;
that cost is acceptable in a reference that teaches ownership and CQRS.

We avoid a folder per individual use case until it has several supporting files.
Commands and queries are grouped under the application role with business-named
files. Naming conventions follow each language while preserving the same
conceptual structure across all six contexts.

Compiler/AST checks enforce query pairing and one reaction per file, alongside
existing mutation and dependency rules. Mapping tests cover revisions, missing
rows, continuation identities, corrupt state and retryable storage failures.
Published contracts, durable identities and transaction semantics retain their
existing definitions. Full infrastructure and browser verification still apply.
