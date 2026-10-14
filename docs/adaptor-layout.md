# Browse the external boundaries

Start with a business context to understand what the service does. Its adaptor
folder then explains how commands, queries and events cross its boundaries and
how its own aggregates and views are stored.

| Service                 | Boundary map                                                         | Business owners                                         |
| ----------------------- | -------------------------------------------------------------------- | ------------------------------------------------------- |
| Storefront (Go)         | [Adaptors](../services/storefront/adaptors/README.md)                | Menu, Ordering                                          |
| Operations (Python)     | [Adaptors](../services/operations/src/operations/adaptors/README.md) | Preparation, Collection                                 |
| Engagement (TypeScript) | [Adaptors](../services/engagement/src/adaptors/README.md)            | Loyalty, Communication                                  |
| API (Go)                | [Adaptors](../services/api/adaptors/README.md)                       | HTTP translation, sessions and authorised subscriptions |
| Web (TypeScript)        | [Adaptors](../services/web/src/adaptors/README.md)                   | Browser HTTP and projection decoding                    |

## Read a context adaptor folder

- `messaging/` translates owner requests into application DTOs, maps query replies,
  encodes private facts and declares durable subscriptions. Business resource names
  distinguish files such as `ticket_requests.py` and `account-query-mapping.ts`.
- `persistence/` contains named read/write repositories, snapshot restoration,
  aggregate transaction bindings and owner migrations. Go uses a `postgres/`
  package beneath it.
- `delivery/` implements an outbound business port where needed. Communication's
  notification provider uses the stable notification identity for idempotency.

Query endpoints invoke named query handlers; persistence maps application views.
Aggregate saves stage private facts. Delivered private facts either enqueue an
owner command or independently publish an integration event through a publication
handler. Find the named reaction in the context's application folder and its
subscription binding under `apps/`.

## Read a shared adaptor folder

Persistence owns transactions, exact stored bytes, receipts, outcomes and leased
dispatch. Messaging owns envelope validation, broker consumption, publisher
confirms and relays. Realtime owns browser projection encoding and gateway I/O.
Operational HTTP serves health and authenticated diagnostics. The central command
executor binds a plain handler to one invocation transaction.

Shared boundary code receives owner mappings and generated registrations from
composition. It contains technical mechanisms used by the contexts. The API and
web have their own roles in the table above and use published contracts.

[Decision 0021](decisions/0021-service-and-context-adaptor-layout.md) records the
layout and its executable rules. Native tests sit beside Go/TypeScript subjects;
Python context tests mirror their source paths under `tests/contexts/`.
