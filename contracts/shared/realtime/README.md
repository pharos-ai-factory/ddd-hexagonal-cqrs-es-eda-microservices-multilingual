# Café realtime messages, version 1

[v1/realtime.proto](v1/realtime.proto) defines
the shared `Publication` envelope. It imports typed snapshots owned by each
context and a typed `OperationRejected` outcome, separately from delivered domain
and integration events.

Use this contract when changing the update applied by a browser after a root
transition. Context ownership follows the HTTP contract:

| Context                | Source                                                                                         | Snapshot messages            |
| ---------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------- |
| Menu                   | [menu_snapshots.proto](../../menu/realtime/v1/menu_snapshots.proto)                            | `Drink`, `Edition`, `Offer`  |
| Ordering               | [ordering_snapshots.proto](../../ordering/realtime/v1/ordering_snapshots.proto)                | `Order`, `Line`, `Selection` |
| Preparation            | [preparation_snapshots.proto](../../preparation/realtime/v1/preparation_snapshots.proto)       | `Ticket`                     |
| Collection             | [collection_snapshots.proto](../../collection/realtime/v1/collection_snapshots.proto)          | `Pickup`                     |
| Loyalty                | [loyalty_snapshots.proto](../../loyalty/realtime/v1/loyalty_snapshots.proto)                   | `Account`, `Grant`, `Reward` |
| Customer Communication | [communication_snapshots.proto](../../communication/realtime/v1/communication_snapshots.proto) | `Notification`               |

The filename includes its owning context and snapshot purpose. The files retain
the existing `cafe.realtime.v1` package, field numbers and message names.
Generated bindings remain service-local.

Centrifugo connect and refresh callbacks are HTTP operations defined in
[the API HTTP contract](../../services/api/http_api/realtime.paths.json). Delivered
business facts such as `OrderPlaced` belong to [events](../messaging/events.md).

The envelope identifies the owning context, kind, root, revision, publication and
`DeliveryTarget`. The snapshot ID must match the envelope root ID. A `context`
target delivers shared owner state on `cafe:<context>`. A `principal` target
delivers `OperationRejected` to `principal:<authenticated-subject>`. Applications
choose typed targets; relays never construct transport channels.

Publish explicitly through the owner's `RealtimeBusPort`. Command handlers stage
exact bytes and dispatch intent in their root/outcome transaction. Declared domain
or integration subscribers commit their browser bytes with a receiving receipt.
A delayed subscriber reads the current owner view and uses that view's actual
revision. Keep the original source-event revision separate from the snapshot revision.

`OperationRejected` carries command ID, correlation ID, stable code and a safe
message. Its target must match the authenticated command actor and its revision
can be zero for a missing root. Display it separately from
snapshot state; preserve uncertain command recovery through the recorded reply.
Snapshot publications require a positive revision.
The relay may deliver duplicates or different roots out of order. Consumers apply
only a newer revision for the same root. Centrifugo recovery positions belong to
the whole context stream, not to an individual root.

Generated Go/Python bindings and TypeScript JSON descriptors are committed inside
the owning service's transport ring. Run `pnpm generate:contracts` after schema
changes. Do not reuse field numbers; reserve removed fields. A breaking change
requires a version and explicit client/server transition.

See [the explicit publication recipe](../../../docs/education/concepts/realtime-events.md),
[decision 0026](../../../docs/decisions/0026-explicit-bus-ports-and-realtime-publication.md)
and [CLIENT-SUBSCRIPTIONS.md](../../../CLIENT-SUBSCRIPTIONS.md) for authority, recovery and browser evidence.
