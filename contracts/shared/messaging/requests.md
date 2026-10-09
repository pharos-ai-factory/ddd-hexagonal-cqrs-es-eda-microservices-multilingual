# Commands, queries and replies

The Go API translates its OpenAPI HTTP boundary into versioned Protobuf requests
and sends them to the owning context through RabbitMQ. An owner-side adaptor
translates the wire payload into a plain application command or query. Domain
and application packages keep their existing transport-independent types.

```text
HTTP request → API boundary translation → Protobuf request → RabbitMQ
    → owner boundary translation → application handler
    → Protobuf reply → RabbitMQ → API HTTP response
```

## Finding a contract

Start at `contracts/<context>/messaging/v1/<context>_requests.proto` for the
owner's request, command, query and reply envelopes. Their Protobuf packages are
`cafe.<context>.requests.v1`. [common.proto](v1/common.proto) defines shared
command metadata, pagination, outcomes and technical failures.
[validation.proto](v1/validation.proto) defines the explicit `required_input`
field annotation. Shared messaging contains technical types only.

Each domain service generates request bindings for its own contexts. The API
generates all six. API and owners dispatch using native context envelopes.
Generated Go interfaces and factories select envelopes by owner. Payload field
names and numbers remain local to each context. An owner schema compiles with
its own payload files and shared technical definitions.

Under `contracts/<context>/messaging/`, each owner publishes these sources:

| File | Responsibility |
| --- | --- |
| `v1/<owner>_requests.proto` | Owner envelopes and their payload tags |
| `commands/v1/<owner>_commands.proto` | Explicit command input messages; present only for contexts with externally callable commands |
| `queries/v1/<owner>_queries.proto` | Named list and item queries; optional page input |
| `queries/v1/<owner>_replies.proto` | Owner query DTOs, loaded versions and list/page results |

For example, Menu owns `CreateDrink`, `PublishEdition`, `GetDrink` and `ListEditions`.
Communication currently exposes queries; its event-triggered application commands
remain internal. Add a wire contract when a command becomes a published interface.
Query DTOs evolve independently of browser snapshots and delivered event payloads.

Command fields use explicit scalar presence so zero prices and versions remain
distinguishable from missing input. Required fields carry `(cafe.requests.v1.required_input) = true`; an unannotated
optional field stays optional. Boundaries reject missing required fields,
foreign owners, invalid identities and unsupported envelope versions before
invoking an application handler. Message names and field numbers become immutable
once published; additive evolution requires compatible defaults or a new version.

## Delivery and uncertainty

The bootstrap owns durable quorum queues and the topic exchanges `cafe.requests`
and `cafe.replies`. Each context has independent `ref.<owner>.commands` and
`ref.<owner>.queries` queues and consumers and publishes `reply.<owner>`.
The API publishes `request.<owner>.command` or `request.<owner>.query` and exclusively consumes
`ref.api.<owner>.replies`. Runtime credentials cannot declare, bind or delete
resources. Topic permissions prevent owners from sending requests or forging
another owner's reply. The API cannot publish domain/integration events.

The current development composition has one API process, with concurrent callers
matched by a fresh transport request ID. Its exclusive reply consumers enforce
that deployment constraint. Multiple API processes would require a reviewed
reply-addressing and queue-ownership design.

Requests and replies use persistent, mandatory publication with publisher
confirms. Command consumers acknowledge after their receiving transaction commits
exact reply bytes in `cafe.command_replies` and a publication intent in
`cafe.command_reply_dispatches`. The same transaction commits aggregate state,
command receipt, outcome and outgoing event/realtime intent. A separate publisher
recovers reply delivery using leased, fenced dispatch rows and confirmed mandatory
publication. Runtime credentials can insert and read reply bytes; they cannot
rewrite them. Reply encoding or SQL failure rolls back the whole transaction.

Reply intents have a thirty-second lifetime, which bounds publication retries.
Eight-second leases allow abandoned claims to recover within that lifetime.
Expired replies retain their evidence and completed dispatch reason. Explicit
command retries use a fresh transport request ID, recover the original outcome
from the stable command receipt and atomically append a fresh reply intent.
Queries publish a confirmed reply before acknowledgement. Separate consumers
allow queries to read committed state while a command waits for its owner lock.
Invalid deliveries enter the matching owner/kind dead-letter queue. Quorum
delivery limits bound repeated connection-failure redelivery. Infrastructure
failures produce a technical unavailable reply and allow an explicit retry.

Commands retain the supplied command ID, aggregate ID, expected version and
workflow correlation ID. Transport request IDs change on retries and match replies;
they never become command deduplication keys. Query adaptors construct named
application query DTOs; their handlers use application-owned read models and
named reader ports. Pagination sends an owner identity cursor; the API translates that
into the resource-bound opaque HTTP cursor.

The API waits up to twelve seconds. A timeout or lost connection returns HTTP
503 and leaves command completion uncertain. The browser retains the original
attempt for retry. Broker confirmation proves publication, while a successful
command reply reports the committed owner outcome. Request queues expire waiting
messages after fifteen seconds; expiry cannot cancel a handler already running.
Replies expire after thirty seconds; late and duplicate replies are acknowledged
and discarded when no caller remains.

Direct owner HTTP endpoints remain available for development inspection and their
own OpenAPI conformance. API business dispatch uses RabbitMQ exclusively.
Centrifugo session callbacks, realtime publication and the development notification
provider retain their dedicated technical transports.

## Verification

Run `pnpm generate:contracts` and review service-local bindings. `pnpm verify`
checks OpenAPI-to-Protobuf translation for all 31 business operations and native
command/query adaptors. `pnpm test:integration` exercises the actual request/reply
path through Go, Python and TypeScript, denied broker operations and recovery of
a committed command after its reply is deliberately made unroutable.
