# Browser delivery contract

The Next.js frontend uses Centrifugo with typed Protobuf realtime messages. The separate
Go API owns authentication and subscription authorisation. Domain services own
business state, owner queries and durable realtime publication intents.

## Session and channel authority

`POST /auth/login` accepts the local operator access code only from a configured
Origin. The API provisions a fixed operator principal, stores an opaque session
under a hashed key in Valkey, and returns an HttpOnly, SameSite=Strict cookie.
Sessions expire after one hour; HTTPS origins receive Secure cookies.

Centrifugo calls the internal `/api/realtime/connect` proxy with a dedicated
secret and the forwarded Cookie and Origin. The API authenticates that session
and selects seven channels: one per context (`cafe:menu`, `cafe:ordering`,
`cafe:preparation`, `cafe:collection`, `cafe:loyalty`, `cafe:communication`),
plus `principal:<authenticated-subject>` for private command outcomes.
Browser-supplied channel choices are ignored. Client subscribe and publish
commands are disabled in Centrifugo. The JavaScript client uses
`centrifuge/build/protobuf` and no `newSubscription` calls.

This operator may inspect all six contexts, including collection codes. The
customer identity entered on screen is a domain subject for orders, not an
authenticated identity or an authorisation claim. This development simplification
does not implement production identity, tenancy or resource policies.

Logout removes the session and atomically records durable disconnect work.
Expiry scanning uses a persistent deadline index. Each revoked session has its
own work identity, preventing an older worker from deleting a later revocation.
The API retries Centrifugo disconnection until accepted. All operator connections
are disconnected; any other valid session must authenticate again on reconnect.
There is no single-connection or single-window setting.

Connect responses carry a one-second authorisation deadline, calculated before
reading the session. After observing durable revocation, the worker records a
deadline in Valkey and waits beyond that grant window before its final disconnect.
It crosses the whole deadline second because Centrifugo accepts an equal
`expire_at`. A successful response still in flight at completion is therefore
expired when Centrifugo receives it. Worker recovery retains pending work and
its deadline; an expired deadline key safely starts another wait.

Centrifugo refreshes established connections through the internal
`/api/realtime/refresh` proxy, forwarding their original Cookie and Origin.
The API checks that session again and grants at most another minute. Refresh
does not create a connection or restore a disconnected one. These checks do
not query business state or create browser polling.

## Atomic publication

A command explicitly publishing through `RealtimeBusPort` commits its root,
version, outcome/receipts, private facts and **exact Protobuf browser bytes** together.
A save alone produces no browser message. Declared domain/integration event handlers
commit their receiving receipt and outgoing browser bytes in a separate transaction. The browser contract is `contracts/shared/realtime/v1/realtime.proto`.
It is separate from the broker's domain/integration envelope.

Immutable `realtime_publications` and mutable `realtime_dispatches` are separate.
A relay claims a fenced lease and sends `b64data` with the publication identity
as Centrifugo's `idempotency_key` through the internal realtime gateway. Each
context has a distinct publisher credential. Relays do not send a channel. The
gateway validates the credential, message owner, payload and typed delivery target,
then derives the Centrifugo channel while forwarding the original Protobuf bytes. Only
the gateway and Centrifugo receive the full server API key. The session worker
uses a separate credential authorising disconnection only. A relay completes
its dispatch after acceptance.
Failure leaves retryable work. A publication accepted before a lost response may
be sent again; the browser's revision guard independently makes duplicates safe.

Each publication carries an event ID, contract version, owning context, aggregate
kind/ID, monotonic root revision, typed delivery target and one explicit payload.
The decoder checks the channel, target, context, kind, identity and revision before
dispatching it. Context-targeted messages carry authoritative snapshots. A delayed event handler reads
an owner view and uses that view's actual revision. `OperationRejected` messages carry
command/correlation identity and a stable rejection independently of snapshot ordering.
They target the authenticated command principal and may carry revision zero when
the root is absent. The frontend displays them
without changing aggregate state or clearing uncertain command attempts.

## Revisions, history and reconciliation

Centrifugo epoch/offset describes a context channel containing many roots. The
application revision belongs to one aggregate, keyed by kind and ID. These are
separate ordering mechanisms. A later root snapshot wins; an older query response
or duplicate publication cannot roll it back.

Each window owns its connection and transport cursors:

1. Attach all seven server-selected subscriptions before beginning reconciliation.
2. On initial attachment, traverse every page of each of the eight resource lists.
3. Apply incoming snapshots by root revision while queries are in flight.
4. On a successfully recovered subscription, consume history without business GETs.
5. On any unrecoverable subscription, traverse the eight owner lists once for
   that connection. A traversal can require several requests. Query failure is
   visible and an explicit reconnect retries it.

There is no periodic business polling. A disconnect triggers one session check
where appropriate; this is access revalidation, not projection polling. Pending
business steps remain visible, for example an earned grant awaiting Reward
issuance. Browser commands are disabled while the transport is reconnecting.

No root deletion exists in this model. Deletion would require versioned tombstones
and reconciliation removal semantics. Queries support optional UUID keyset
pagination with resource-bound opaque cursors. The browser requests 100 items
per page until `nextCursor` is null. Live subscriptions cover concurrent insertions
behind the cursor, and root revisions protect against stale pages. A traversal
does not claim a single database snapshot across requests. Ordinary queries can
also return their complete unpaginated result; pagination is an explicit port
capability, not a requirement on every query.

## Command acceptance and uncertainty

All browser commands go through the Go API, which translates HTTP JSON into
Protobuf requests over RabbitMQ. Owner adaptors construct the application command.
The API waits for its committed outcome; timeout preserves uncertainty. Before sending, a window persists
its body, root version, command ID and correlation ID in session storage. A
network failure or server failure retains that attempt for an explicit identical
retry, including after page reload. An authentication failure also retains the
attempt and opens sign-in; after authentication, retry uses the same saved
body, version and identities to recover the outcome. A known rejection clears the attempt; a new
business attempt uses a new ID. Success does not optimistically invent downstream
state: Centrifugo or an authoritative reconciliation supplies the current state.

## Evidence

`scripts/browser/cafe.spec.ts` exercises two independent windows, a lost command
response, the entire three-order workflow, no workflow polling, recoverable
history, deliberate history loss, mobile layout and logout propagation.
Go API tests cover authentication, origin checks, channel authority and
HTTP-to-Protobuf request translation. The broker reply-loss fixture verifies
recovery of committed commands with their original identities. Each runtime's PostgreSQL lane proves atomic realtime
intent and lease fencing. `pnpm test:integration` includes the real browser lane.

## Working on the browser

Feature folders under `services/web/src/features/` follow user tasks: menu,
ordering, preparation, collection, rewards and notifications. The page composes
these views; business invariants remain in their owning services.

Use `adaptors/http/operation-client` and its generated operation/request/response types for
HTTP. Change the owning OpenAPI source and regenerate before adjusting a caller.
Keep command persistence/retry in `shared/commands/`, message validation and
reconciliation in `shared/realtime/`, and common formatting in `shared/ui/`.
`RealtimeProvider` owns the Centrifugo connection. Features register typed handlers
with `useRealtimeMessage(realtimeMessages.<name>, handler)` and read status through
`useRealtimeConnection`. Hooks never create subscriptions or name channels.

A feature change must retain the saved uncertain command, subscription barrier,
per-root revision guard and explicit reconnect behaviour. Update the browser
realtime contract separately when its public shape changes. Native frontend
tests check mapping/reconciliation; the full integration lane verifies independent
windows, uncertain-command recovery and in-flight session revocation.

Integration publication has its own private subscriber receipt/outbox transaction.
Ordering and Preparation have separate domain-to-realtime subscriptions; Loyalty
explicitly subscribes to `RewardIssued` integration events for its issuance update.
Other commands publish selected updates within their transaction. These paths have
different delivery timing and share revision-aware recovery. See
[decision 0026](docs/decisions/0026-explicit-bus-ports-and-realtime-publication.md).
