# 0031 — Targeted realtime messages

Status: accepted, 2026-10-22.

## Context

Decision 0026 made browser publication explicit, but its application value carried
only a snapshot or rejection. Persistence selected `cafe:<context>` directly and
relays sent that channel to the gateway. The browser connection and state handling
also lived in one café-specific hook. These choices did not express recipient
authority in the message contract and did not provide a reusable publisher or
frontend handler API.

The café needs two delivery audiences. Current owner state is visible to every
operator authorised for that context. A recorded command rejection belongs to the
authenticated operator who invoked that command.

## Decision

Every realtime publication carries a typed application `RealtimeTarget`:

- `context:<owner>` carries an authoritative state message shared with authorised
  members of that context;
- `principal:<subject>` carries `OperationRejected` for one authenticated actor.

Go, Python and TypeScript expose `RealtimeMessage`, context/principal target
factories, snapshot/rejection factories and `RealtimeBusPort`. Context handlers
choose targets without importing Centrifugo concepts. Command persistence requires
a principal rejection target to equal the command's authenticated actor. Internal
commands without an actor do not publish a private rejection.

The Protobuf `Publication` envelope adds `DeliveryTarget` at field 7. Owner adaptors
encode the selected target and persist the derived diagnostic channel. Relays send
only the immutable Protobuf bytes and publication identity. The gateway validates:

- the publisher credential matches the envelope owner;
- the payload alternative belongs to that owner;
- context targets match the owner and carry state;
- principal targets contain a canonical subject and carry `OperationRejected`.

The gateway derives `cafe:<owner>` or `principal:<subject>` after validation.
Relays omit `channel`. The existing publisher request field remains as an optional,
deprecated compatibility hint; the gateway accepts it only when it equals the
derived channel, and it has no authority to select an audience. Centrifugo clients
cannot publish or subscribe directly.

The session authority selects all six café context channels and the authenticated
principal channel. The browser does not send a subscription list. `RealtimeProvider`
owns the connection, recovery barrier, decoding and recent-message deduplication.
Features register typed handlers with:

```ts
useRealtimeMessage(realtimeMessages.orderChanged, (message) => {
  // message is the typed orderChanged shape
});
```

`useRealtimeConnection` exposes connection health. Café-specific state merging and
query reconciliation remain application consumers of this transport API.

## Consequences and compatibility

The target vocabulary is application-defined. This reference uses `context` and
`principal`; another application can define its own target catalogue without
reusing café or platform domain names.

Adding field 7 is wire-additive, while target presence is required by the new
gateway and browser decoder. Development deployments must drain or discard queued
realtime publications and disposable Centrifugo history created by the earlier
contract before switching. Aggregate state, domain/integration events and command
receipts are unchanged.

This refines decisions 0002, 0023 and 0026. Context projections used for service
reads remain separate concepts; the realtime SDK exposes messages and handlers.

## Verification

Native tests cover target factories, scoped publication buffers, actor matching,
context and principal encoding, server-selected subscriptions and browser decoding.
Gateway tests reject conflicting publisher channel hints, owner/payload mismatches,
malformed targets and invalid principal identities. Integration lanes retain the
existing transaction, relay, history-gap and revocation evidence.
