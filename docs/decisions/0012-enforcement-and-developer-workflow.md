# 0012 — Enforced use-case boundaries and the development workflow

Status: accepted, 9 October 2026. Extends decisions 0009–0011.

## Command execution

Application code cannot acquire another command handler's execution capability.
Driving adaptors and context composition invoke handlers. Event reactions use the
owner durable command port. Each command handler makes at most one direct call to
its aggregate store. Capturing that method, putting it in a loop or nested closure,
or calling it twice is rejected. Domain mutation inside the store's decision
callback remains supported.

The compiler/AST checks deliberately accept a small, explicit implementation shape.
They reject two mutually exclusive store call sites as well: select inputs first
and use one call. These checks complement the database's one-target-per-transaction
guard. They do not claim to prove arbitrary reflective programs. Negative fixtures
cover direct handler chaining, aliases, repeated execution and ordinary mutation.

## Composition and shutdown

Context resources have one explicit lifetime. Failed graph construction, resource
initialisation or subscription registration closes resources already acquired.
Workers stop and drain before their pools close, including RPC request handlers.
Python uses an ExitStack, TypeScript an application lifecycle scope, and Go a
resource scope surrounding Fx startup/shutdown. Fx still owns the runtime graph.
Each language has deterministic failure-path tests.

A context's subscription manifest must correspond to exactly one incoming-event
consumer and one command consumer per declaration. Missing, duplicate, foreign
and mismatched bindings fail composition. Typed event definitions couple incoming
names to payload decoders; application handlers retain ordinary constructor ports.

## Historical compatibility

`pnpm check:compatibility` compares compiled public Protobuf descriptors, private
stored formats and bundled OpenAPI documents with an accepted Git revision.
Local verification defaults to commit `caa3206`, the initial durable-command
baseline. `--against <commit>` selects another baseline. CI fetches history and
uses the PR base SHA or the preceding main commit through `CAFE_CONTRACT_BASE`.
Baseline source is read as data and is never executed.

Existing Protobuf message, enum and field identities and semantics must remain
stable. Optional additions are supported; changed custom field semantics require
review. OpenAPI checks use request/response direction: new required inputs and
weakened response guarantees fail. Unknown changed schema constraints fail
conservatively. This policy may reject a compatible refactoring; evaluate it
explicitly and add precise rules and negative tests rather than a broad bypass.
An intentional incompatible public interface gets a new version while the old
version remains supported for its consumers.

Historical private command fixtures cannot be edited or removed relative to the
baseline. Native codec tests decode those bytes and compare their meaning.
Queued private work retains storage compatibility without becoming a published
integration contract. Semantic changes still require domain and scenario review.

## Readiness and inspection

`pnpm dev:status` checks database/session diagnostics, live RabbitMQ consumer
registrations and the realtime health endpoint. `dev:up` waits on the same
readiness policy. Explicitly paused
subscriptions omit their two consumers; owner request consumers remain required.
`/healthz` retains its liveness meaning. Readiness does not require an empty queue.

`pnpm workflow:inspect <uuid>` reads command receipts, acceptance records, fenced
dispatch status and committed outcomes across the six development databases.
It follows discovered correlation IDs and displays safe envelope identities.
It never consumes, requeues or acknowledges messages. Dead-letter depths are
queue-wide evidence, independently labelled; an absent outcome cannot establish
which queue currently holds an individual message. Results are bounded and
explicitly report truncation. Payloads and credentials are omitted.

## Developer tools

Focused tests and automatic Compose rebuilds shorten the edit/test loop. Full
`pnpm verify` and `pnpm test:integration` remain mandatory for hand-off.
Scaffolding creates reviewable files in `.local/scaffolds`, including role names,
port shapes, registration guidance and deliberately failing behaviour tests.
Developers define domain policy and copy the reviewed pieces into their owner.
The generator cannot infer invariants, event mappings or safe Protobuf field tags.
Debug configurations target native tests without opening public debug listeners.
Formatting operates on explicit files or changed tracked files and leaves generated
bindings to their generators. All runtime tooling remains development-only.
