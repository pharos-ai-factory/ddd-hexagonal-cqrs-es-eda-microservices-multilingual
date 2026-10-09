# Integration event contracts

Shared event contracts describe facts published to another bounded context.
Find a producer's payloads under
`contracts/<context>/messaging/integration_events/v1/`. Each publishing context
owns its `catalogue.json`, including the published name, producer aggregate and
consumer binding. The technical envelope is [v1/events.proto](v1/events.proto).

| Producer | Published events |
| --- | --- |
| Menu | `menu.edition-published` |
| Ordering | `ordering.order-placed` |
| Preparation | `preparation.drinks-ready` |
| Collection | `collection.pickup-opened`, `collection.order-collected` |
| Loyalty | `loyalty.reward-issued` |

Communication currently publishes no integration event. Its notification delivery
work is private to Communication.

## Private domain facts and delivery

Domain facts are plain, context-owned types. Their definitions have no public
contract version. An internal reaction can use RabbitMQ and the transactional
outbox without becoming an integration contract.

The current internal deliveries are Menu's drink directory, Loyalty's reward
issuance and Communication's notification dispatch. Their plain application
values, private Protobuf delivery formats, fixtures and queue metadata live in
the owning service context. They are excluded from shared payloads and catalogues.
The deployed format retains its historical metadata field numbers and exact
payload bytes so stored outbox records and queued work can recover. That local
storage compatibility does not establish a published interface.

Private and integration messages retain the same durable delivery mechanism:
atomic aggregate/receipt/outcome/outbox, publisher confirms, manual acknowledgements,
bounded retries, dead-letter preservation and explicit replay. Aggregate-changing
reactions first commit a durable owner command; the command consumer performs the
root transaction. Projection handlers update only their read data. See
[the subscription recipe](../../../docs/developer-workflow.md#wire-an-aggregate-changing-subscription).
Private topics
remain restricted to owner consumers. The deployment bootstrap gathers internal
queue metadata independently of the published contract catalogue.

## Public envelope and broker metadata

The envelope identifies the event, producer context, published contract version,
source aggregate kind/ID/version, correlation, causation and occurrence time.
The typed payload must agree with the published catalogue. Removed private
payload numbers and names are reserved in the public envelope.

The event ID identifies one immutable publication. Correlation groups a workflow;
causation identifies the producing command. A consumer receipt retains the
original received bytes for its duplicate fingerprint.

AMQP uses `cafe.events`, routing key `integration.<event-name>`, persistent
`application/x-protobuf` messages and publisher confirms. Message ID, type,
application ID and correlation ID match the envelope. The `contract-version`
header equals its version. Owner-private topics use `domain.<event-name>` as
an internal delivery convention.

## Evolution and evidence

Preserve public message names and field meanings. Never reuse a field number;
reserve removed numbers and names. A breaking public change needs a reviewed
version transition. Consumer codecs validate producer authority, expected payload,
identities and published evidence before invoking application handlers.

The public compatibility fixture is
[RewardIssued](../../loyalty/messaging/integration_events/v1/fixtures/reward-issued.v1.hex).
Go verifies exact encoding and round trip; Python and TypeScript decode the same
published message. Owner-local fixtures separately prove that the relocated
private Menu and Loyalty formats preserve stored bytes.

The workflow lane exercises every public integration event and every private
reaction through real PostgreSQL and RabbitMQ. Acceptance tests retain private
payloads as opaque bytes when introducing a fresh delivery ID; they do not load
another context's private message schema.

Run `pnpm generate:contracts`, `pnpm verify` and `pnpm test:integration`.
Bindings stay in each consuming service; domain and application packages use
plain types. Protobuf import identities are mapped by `scripts/contract_sources.py`.
Replay restores original delivery identity and bytes after confirmed publication;
current-state persistence remains authoritative.
