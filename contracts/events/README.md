# Delivered event contracts

The canonical wire schema is [proto/cafe/v1/events.proto](proto/cafe/v1/events.proto).
Generated Go/Python types and TypeScript descriptors stay in transport/adaptor
rings. Applications exchange plain DTOs; domains use their own facts.

## Catalogue

| Event name | Scope | Producer aggregate | Consumer |
| --- | --- | --- | --- |
| `menu.drink-published` | Private domain | Drink | `menu.drink-directory` |
| `menu.edition-published` | Integration | MenuEdition | `ordering.menu-directory` |
| `ordering.order-placed` | Integration | Order | `preparation.accept-order` |
| `preparation.drinks-ready` | Integration | PreparationTicket | `collection.open-pickup` |
| `collection.pickup-opened` | Integration | Pickup | `communication.pickup-notice` |
| `collection.order-collected` | Integration | Pickup | `loyalty.credit-collection` |
| `loyalty.reward-earned` | Private domain | LoyaltyAccount | `loyalty.issue-reward` |
| `loyalty.reward-issued` | Integration | Reward | `communication.reward-notice` |
| `communication.notification-requested` | Private domain | Notification | `communication.deliver-notice` |

`catalogue.json` records shared ownership metadata. Go topology and codecs use
the corresponding service-local `contracts/events/model/catalogue.go`; Python and
TypeScript use generated catalogue copies. The architecture gate compares them. A private event is not a public
integration contract merely because RabbitMQ carries its bytes.

## Envelope and broker metadata

Every message identifies its event, producing context, visibility, contract
version, source aggregate kind/ID/version, correlation, causation and occurrence
time. The typed Protobuf `oneof` must agree with the name and catalogue.

The event ID identifies one immutable publication. Correlation groups a workflow;
it never deduplicates it. Causation identifies the producing command. For
event-triggered commands, the command receipt additionally records its incoming
causation, allowing the chain to be followed across transactions.

AMQP uses:

- Exchange: `cafe.events`, a durable topic exchange.
- Routing key: `<domain|integration>.<event-name>`.
- Content type: `application/x-protobuf`; persistent delivery mode.
- `MessageId`, `Type`, `AppId`, `CorrelationId`: copies of matching envelope fields.
- `contract-version`: an AMQP 32-bit integer equal to the envelope version.

The consumer validates the bytes, payload meaning and AMQP metadata before
calling a handler. No events are routed by Go package names.

For consumer `loyalty.issue-reward`, queues are
`ref.loyalty.issue-reward`, `.retry` and `.dead`. Retry and replay use the
context-owned direct exchange `ref.loyalty.delivery`. See the
[delivery decision](../../docs/decisions/0001-consistency-and-delivery.md).

## Evolution and cross-language fixtures

Codecs accept version 1 and reject unsupported versions, wrong producers/scopes
and mismatched payloads. Go additionally rejects unknown envelope fields; Python
and TypeScript follow their Protobuf libraries' unknown-field handling. Extra
payload fields follow Protobuf's decoding behaviour and are not surfaced to
application DTOs. Do not assume arbitrary schema additions will be accepted by
every deployed decoder.

Never reuse a field number or alter the meaning of stored bytes. Reserve removed
field numbers. For a breaking change, introduce a versioned contract and an
explicit transition with old and new decoders/consumers. Private persisted
messages need the same care even if only one context consumes them.

`fixtures/reward-earned.v1.hex` contains a fixed private-domain publication.
`services/storefront/contracts/events/protobuf/codec_test.go` proves exact Go encoding and semantic round trip,
and rejects invalid contract combinations. Python and TypeScript tests decode the same fixture. All consumers preserve
original bytes during retry/replay.
Protobuf serialisation need not be byte-identical across every implementation;
the consumer's duplicate fingerprint describes the received publication,
not a newly re-encoded interpretation.

The complete six-context HTTP journey exercises every delivered event type.
Add language-neutral fixtures and negative cases as a new implementation is
introduced; the one initial golden message is not a complete compatibility suite.

## Generation

```sh
pnpm generate:contracts
pnpm verify
```

The script pins `grpcio-tools` and `protoc-gen-go`. Generated source is committed
so normal service builds do not require Protobuf code generation tools. Review
generated diffs alongside their schema and fixture changes.
Python `.pyi` declarations are generated alongside the runtime bindings. Mypy
uses them in the adaptor ring; application DTOs remain independently defined.

## Replay is delivery recovery

The replay executable moves one original dead-letter message back to its
consumer after confirmed publication. It preserves ID, content and envelope.
It resets retry-attempt metadata, not business state or receipts. Fix a poison
message's cause before replaying it.

Replay is not event sourcing, deletion of deduplication records, or permission
to issue a new business action under an old event identity.
