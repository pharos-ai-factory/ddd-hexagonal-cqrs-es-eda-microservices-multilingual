# Shared Operations adaptors

These packages implement the technical boundaries used by Preparation and
Collection. Each context keeps its business mappings, private schemas,
subscriptions, migrations and named repositories under its own `adaptors/`.
Start with [Preparation](../contexts/preparation/README.md) or
[Collection](../contexts/collection/README.md) to find a business capability.

## Find a responsibility

| Location                                     | Responsibility                                                                                                             |
| -------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| [command_execution.py](command_execution.py) | Invoke a named handler with its scoped repository and collect selected private facts                                       |
| [persistence](persistence)                   | PostgreSQL pools, aggregate transactions, snapshot repositories, receipts, exact outgoing bytes and fenced dispatch leases |
| [messaging](messaging)                       | Protobuf envelopes, RabbitMQ consumers, request dispatch and confirmed outbox/reply relays                                 |
| [realtime](realtime)                         | Browser projection encoding and publication through the authorised realtime gateway                                        |
| [http](http)                                 | Authentication and error handling for health/diagnostics HTTP                                                              |
| [diagnostics.py](diagnostics.py)             | Database backlog evidence and redacted worker failures                                                                     |
| [generated](generated)                       | Generated Protobuf bindings, event/request registries and persistence metadata                                             |

The [composition root](../apps/service.py) chooses resources and starts workers.
The [subscription helper](../apps/runtime.py) binds durable reactions and validates
the complete context declaration. Package initialisers contain documentation only;
callers import the concrete module for the capability they need.

## Follow a completed ticket

1. Preparation's `CompletePreparationCommandHandler` loads its `PreparationTicket`,
   invokes `complete` and saves through the named write repository. The root moves
   from preparing to ready and queues a plain `DrinksReadyDomainEvent`.
2. The context's [ticket mapper](../contexts/preparation/adaptors/messaging/ticket_publications.py)
   selects its private delivery. The shared aggregate transaction commits the root,
   receipts, outcome, private event bytes, realtime bytes and any API reply together.
3. The RabbitMQ outbox relay claims a database lease, publishes the stored private
   bytes and completes the lease after publisher confirmation.
4. Preparation's [publication handler](../contexts/preparation/application/event_handlers/publish_drinks_ready.py)
   maps the delivered fact to its integration DTO. The integration-event outbox
   commits the subscription receipt, exact integration bytes and dispatch intent
   together. The consumer then acknowledges the private delivery.
5. The outbox relay publishes those integration bytes with a separate confirm.
   Collection can accept that event and durably enqueue `OpenPickup` in its own
   database. Its command consumer executes the aggregate transition afterwards.

A ready ticket awaiting integration publication is a permitted pending state.
The realtime gateway relay independently publishes the browser bytes committed
with the root. Retries preserve stored bytes and identities. Database commits,
consumer acknowledgements and publisher confirmations have distinct roles.

## Extend the boundary

Keep database access and lease updates in `persistence/`; network delivery belongs
in `messaging/` or `realtime/`. Relays call the persistence capabilities. Shared
codecs use generated registries; context adaptors perform business-specific
validation and restoration. Request envelope factories are generated from the
service's declared request families.

Regenerate with `pnpm generate:contracts` after changing a generator or schema.
Preserve historical wire names, queued bytes, receipt identities and migration
checksums. See [decision 0020](../../../../../docs/decisions/0020-operations-adaptor-layout.md).
