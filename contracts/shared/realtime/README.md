# Café browser projections, version 1

[v1/realtime.proto](v1/realtime.proto) defines
the shared `Publication` envelope. It imports typed snapshots owned by each
context, separately from delivered domain and integration events.

Use this contract when changing the update applied by a browser after a root
transition. Context ownership follows the HTTP contract:

| Context | Source | Snapshot messages |
| --- | --- | --- |
| Menu | [menu_snapshots.proto](../../menu/realtime/v1/menu_snapshots.proto) | `Drink`, `Edition`, `Offer` |
| Ordering | [ordering_snapshots.proto](../../ordering/realtime/v1/ordering_snapshots.proto) | `Order`, `Line`, `Selection` |
| Preparation | [preparation_snapshots.proto](../../preparation/realtime/v1/preparation_snapshots.proto) | `Ticket` |
| Collection | [collection_snapshots.proto](../../collection/realtime/v1/collection_snapshots.proto) | `Pickup` |
| Loyalty | [loyalty_snapshots.proto](../../loyalty/realtime/v1/loyalty_snapshots.proto) | `Account`, `Grant`, `Reward` |
| Customer Communication | [communication_snapshots.proto](../../communication/realtime/v1/communication_snapshots.proto) | `Notification` |

The filename includes its owning context and snapshot purpose. The files retain
the existing `cafe.realtime.v1` package, field numbers and message names.
Generated bindings remain service-local.

Centrifugo connect and refresh callbacks are HTTP operations defined in
[the API HTTP contract](../../services/api/http_api/realtime.paths.json). Delivered
business facts such as `OrderPlaced` belong to [events](../messaging/events.md).

The envelope identifies the owning context, kind, root, revision and publication.
The snapshot ID must match the envelope root ID. Channel authorisation comes from
the API session, not from these payload fields. Each context uses `cafe:<context>`.

Store exact encoded bytes and dispatch intent in the root transaction. Never
re-read current state later to fabricate a publication for an earlier revision.
The relay may deliver duplicates or different roots out of order. Consumers apply
only a newer revision for the same root. Centrifugo recovery positions belong to
the whole context stream, not to an individual root.

Generated Go/Python bindings and TypeScript JSON descriptors are committed inside
the owning service's transport ring. Run `pnpm generate:contracts` after schema
changes. Do not reuse field numbers; reserve removed fields. A breaking change
requires a version and explicit client/server transition.

See [CLIENT-SUBSCRIPTIONS.md](../../../CLIENT-SUBSCRIPTIONS.md) for authority, recovery and browser evidence.
