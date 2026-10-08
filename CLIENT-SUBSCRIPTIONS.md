# Browser delivery contract

The Next.js frontend uses Centrifugo with Protobuf browser projections. The separate
Go API owns authentication and subscription authorisation. Domain services own
business state, owner queries and durable realtime publication intents.

## Session and channel authority

`POST /auth/login` accepts the local operator access code only from a configured
Origin. The API provisions a fixed operator principal, stores an opaque session
under a hashed key in Valkey, and returns an HttpOnly, SameSite=Strict cookie.
Sessions expire after one hour; HTTPS origins receive Secure cookies.

Centrifugo calls the internal `/api/realtime/connect` proxy with a dedicated
secret and the forwarded Cookie and Origin. The API authenticates that session
and selects exactly six channels, one per context: `cafe:menu`, `cafe:ordering`,
`cafe:preparation`, `cafe:collection`, `cafe:loyalty`, `cafe:communication`.
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

A changed root, its version, command outcome/receipts, domain/integration outbox
and **exact Protobuf browser snapshot bytes** commit in the same PostgreSQL
transaction. The browser contract is `contracts/shared/realtime/v1/realtime.proto`.
It is separate from the broker's domain/integration envelope.

Immutable `realtime_publications` and mutable `realtime_dispatches` are separate.
A relay claims a fenced lease and sends `b64data` with the publication identity
as Centrifugo's `idempotency_key` through the internal realtime gateway. Each
context has a distinct publisher credential authorising exactly its own channel.
The gateway validates transport identity and forwards the original bytes. Only
the gateway and Centrifugo receive the full server API key. The session worker
uses a separate credential authorising disconnection only. A relay completes
its dispatch after acceptance.
Failure leaves retryable work. A publication accepted before a lost response may
be sent again; the browser's revision guard independently makes duplicates safe.

Each publication carries an event ID, contract version, owning context, aggregate
kind/ID, monotonic root revision and one explicit typed snapshot. The decoder
checks the channel, context, kind, identity and revision before applying it.
This demonstration publishes snapshots directly from authoritative root state;
it does not expose private domain event payloads as browser messages.

## Revisions, history and reconciliation

Centrifugo epoch/offset describes a context channel containing many roots. The
application revision belongs to one aggregate, keyed by kind and ID. These are
separate ordering mechanisms. A later root snapshot wins; an older query response
or duplicate publication cannot roll it back.

Each window owns its connection and transport cursors:

1. Attach all six server-selected subscriptions before beginning reconciliation.
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
state: Centrifugo or an authoritative reconciliation supplies the projection.

## Evidence

`scripts/browser/cafe.spec.ts` exercises two independent windows, a lost command
response, the entire three-order workflow, no workflow polling, recoverable
history, deliberate history loss, mobile layout and logout propagation.
Go API tests cover authentication, origin checks, channel authority and
HTTP-to-Protobuf request translation. The broker reply-loss fixture verifies
recovery of committed commands with their original identities. Each runtime's PostgreSQL lane proves atomic realtime
intent and lease fencing. `pnpm test:integration` includes the real browser lane.
