# Storefront boundary mechanisms

Menu publishes drinks and immutable priced editions. Ordering builds and places
orders against those editions. Start in their [Menu](../contexts/menu/README.md)
and [Ordering](../contexts/ordering/README.md) context maps for business behaviour.

| Folder                  | Responsibility                                                                                                             |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `command/`              | Bind named command handlers to one aggregate transaction; prepare owner projections before acquiring the write connection. |
| `persistence/postgres/` | Context pools, snapshots, repositories, receipts, outboxes and fenced dispatch leases.                                     |
| `messaging/amqp/`       | RabbitMQ connections, confirms, manual acknowledgements and delivery topology.                                             |
| `messaging/requests/`   | Published request decoding, handler registration and request workers.                                                      |
| `messaging/workers/`    | Event/command outbox relays, consumer recovery and durable command-reply relays.                                           |
| `realtime/`             | Publish committed browser projection bytes to Centrifugo.                                                                  |
| `http/`                 | Operational HTTP responses/security and generated OpenAPI validation.                                                      |

For PlaceOrder, follow Ordering's request mapping to its named command handler,
order write repository and private fact mapper. The shared transaction persists
the order, outcome, reply and dispatch intent together. A private subscription
runs the publication handler before integration consumers see `OrderPlaced`.

[Repository-wide map](../../../docs/adaptor-layout.md).

For RecordCollection, follow the owner-private codec through `CommandOutbox`, the
fenced command relay and the paired consumers registered in composition. The
command handler loads and saves one Order and explicitly publishes its snapshot.
`messaging/internalcommands/` supplies shared identity and receipt validation;
command payload mapping stays in Ordering's messaging adaptor.
