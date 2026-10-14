# Collection adaptors

This boundary owns pickup storage and wire mappings. Start with the
[context map](../README.md) for invariants and named application handlers.

- [Messaging](messaging): resource query endpoints, request/reply mappings,
  private fact codecs, publication mappings and stable subscription declarations.
  Start with [pickup_requests.py](messaging/pickup_requests.py) for the published request boundary.
- [Persistence](persistence): named read/write repositories, snapshot restoration
  and aggregate transaction bindings. [Migrations](persistence/migrations) belong
  to this context's database.
- [Shared service mechanisms](../../../adaptors/README.md): generic transactions, codecs and delivery.

Repository saves stage private domain facts with the root. Subscription handlers
own later durable command hand-offs or integration publication. Pending work is
visible through receipts and dispatch evidence; its meaning is described in the
context map. Query endpoints return application read models through named read
repositories.
