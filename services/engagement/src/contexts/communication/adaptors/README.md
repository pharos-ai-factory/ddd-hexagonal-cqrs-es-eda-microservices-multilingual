# Communication adaptors

This boundary owns notification storage and wire mappings. Start with the
[context map](../README.md) for invariants and named application handlers.

- [Messaging](messaging): resource query endpoints, request/reply mappings,
  private fact codecs, publication mappings and stable subscription declarations.
  Start with [notification-queries.ts](messaging/notification-queries.ts) for the published request boundary.
- [Persistence](persistence): named read/write repositories, snapshot restoration
  and aggregate transaction bindings. [Migrations](persistence/migrations) belong
  to this context's database.
- [Shared service mechanisms](../../../adaptors/README.md): generic transactions, codecs and delivery.

Repository saves stage private domain facts with the root. Subscription handlers
own later durable command hand-offs or integration publication. Pending work is
visible through receipts and dispatch evidence; its meaning is described in the
context map. Query endpoints return application read models through named read
repositories.

[Notification delivery](delivery/http-notification-delivery.ts) implements the outbound provider port with a stable idempotency key.

[Delivery preparation](delivery/notification-delivery-preparation.ts) validates
stored content and obtains the provider receipt before the command registration
helper opens the aggregate transaction. The registration names
`DeliverNotificationCommandHandler` directly.
