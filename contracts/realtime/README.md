# Café browser projections, version 1

`proto/cafe/realtime/v1/realtime.proto` is the browser contract. It contains typed
snapshots for Drink, MenuEdition, Order, PreparationTicket, Pickup, LoyaltyAccount,
Reward and Notification. It is deliberately separate from delivered domain and
integration events.

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

See `CLIENT-SUBSCRIPTIONS.md` for authority, recovery and browser evidence.
