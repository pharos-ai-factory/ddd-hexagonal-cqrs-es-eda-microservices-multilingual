# 0009 — Contract enforcement and durable command replies

Status: accepted, 8 October 2026. Refines request packages and acknowledgement
rules in 0007 and technical sharing in 0008. Addresses the six items in
[contract improvements](../contract-improvements.md).

## Decision

Each context publishes request, command, query and reply envelopes in
`contracts/<context>/messaging/v1/<context>_requests.proto`, with package
`cafe.<context>.requests.v1`. Command inputs and query inputs/results remain in
their respective categories. Shared request schemas contain technical metadata,
pagination, outcomes, failures and validation annotations.

The domain services generate request bindings for their owned contexts; the API
consumes all six. API and owner adaptors encode and decode each context's native
request/reply envelopes, selected by the routing owner. Generated Go interfaces
and factories provide technical dispatch without merging Protobuf oneofs. Payload
names and field numbers are local to their context; separate contexts can reuse
both. Owner envelopes compile independently with shared technical definitions. Existing
field numbers and historical request bytes are preserved by fixture tests. This
development reference intentionally changes the qualified request package names
once to establish ownership; future published package identities remain stable.

Annotate mandatory optional scalar inputs with
`(cafe.requests.v1.required_input) = true`. Boundaries check this annotation,
preserving explicit zero values. Future unannotated optional fields remain
optional. Expected aggregate version remains explicitly required in command
metadata. Owner ACLs construct plain application commands and published query
DTOs through field mappings. Protobuf reflection stays in technical codec/envelope
handling. Application packages remain independent of generated transport types.

The API generates typed HTTP-to-wire mappings from OpenAPI, compiler-produced
Protobuf descriptors (including required-input options) and its
explicit operation/field map. Unsupported operations, unmapped HTTP inputs,
incompatible scalar types and required-input disagreements fail generation.
Frontend operation, input and response types and route metadata are generated
from the same assembled OpenAPI document. Feature code calls a typed operation
client. Architecture checks reject direct frontend fetch calls outside the HTTP
adaptor. Verification mutates real request/response schemas and proves actual
frontend callers fail compilation for incompatible changes. Required headers and
query objects participate in these input types. Shared command and list callers
must satisfy every operation they can select. Fetch supplies the browser's Origin
header for same-origin POSTs; application headers are explicitly typed and sent.
Unsupported parameter locations fail generation.

Each owner has separate command and query queues, routing keys, connections and
consumers: `ref.<owner>.commands`/`request.<owner>.command` and
`ref.<owner>.queries`/`request.<owner>.query`. A command waiting on an owner lock
leaves the query consumer able to read committed state.

For RPC commands, infrastructure scopes a reply encoder around the receiving
transaction without adding transport fields to application metadata. It inserts
exact response bytes and a dispatch intent in the same transaction as the root,
command receipt, outcome and outgoing events/realtime bytes. Duplicate stable
command IDs recover saved outcomes. A fresh transport request ID receives a fresh
reply intent. Reusing a transport identity with different response bytes fails.
Reply encoding and persistence failures roll back all receiving work.

Command consumers acknowledge after this commit. Independent leased, fenced
publishers send the persisted bytes with mandatory publication and confirms.
Publication failure leaves a recoverable intent; a lost confirmation can produce
a duplicate reply. API callers discard late or duplicate replies. Replies have
a thirty-second lifetime and eight-second leases, leaving time to reclaim an
abandoned claim before expiry. PostgreSQL tests in every owner language abandon a
real lease and verify that its original bytes remain deliverable after reclaim.
The lifetime bounds publication retries; expiry records a completed
reason and retains the original bytes. Explicit retries recover the saved command
outcome in a new intent. Queries and technical failures publish confirmed replies
before acknowledgement. HTTP timeouts continue to report uncertain completion.

Owner-local version-two migrations add immutable reply storage and mutable
dispatch tables. Runtime roles have read/insert authority on response bytes and
read/insert/update authority on dispatch rows. Historical bootstrap SQL remains
unchanged. Runtime startup checks migration identities and checksums.

## Verification

`pnpm generate:contracts` regenerates bindings, descriptors, frontend HTTP types,
API mappings and migration metadata. `pnpm check:contracts` regenerates in an
isolated directory and rejects changed, deleted or newly generated files, including
untracked outputs. `pnpm verify` includes this check, native presence/mapping tests,
frontend schema mutation tests and all language compilation/testing gates.
`pnpm test:integration` covers transaction rollback and saved-response recovery in
all owner languages, leased reply fencing, denied response rewrites, independent
command/query consumption and recovery of exact bytes after unroutable publication.
Execution results and environment limits live in [verification](../verification.md).
