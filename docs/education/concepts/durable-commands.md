# Durable commands

A durable command is an owner command whose exact queued representation and
dispatch intent survive a process failure. Event reactions use it to transfer
responsibility before acknowledging the incoming event.

## Creation recipe

1. Name the target command, invariant, aggregate and permitted pending state.
2. Declare the command payload in the owner's `internal_commands.proto` with its `owner` and `visibility: PRIVATE`; add its envelope case. Run generation for its plain input and mechanical mappings. Implement the owner codec's semantic validation and compatibility fixtures.
3. Inject `CommandBusPort` into the event reaction and explicitly map the fact into the owner DTO. Choose stable target/command identities; correlation groups a workflow and does not deduplicate it.
4. Declare the source event, stable subscription identity and private command in `subscriptions.json`. Run `pnpm generate:contracts` for generated identifiers and topology.
5. Bind the concrete durable command outbox, reaction and plain command handler through the shared paired-consumer helper. Preserve separate event-acceptance and command-execution transactions.
6. Test acceptance rollback, repeated delivery, lost confirms/ACKs, retained command bytes and replay. All three languages have executable examples below.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Two independently committed steps

1. The event reaction records receiving evidence, exact command bytes and dispatch
   intent in its local transaction, then ACKs the event.
2. A dedicated command consumer executes the named owner handler, commits its
   aggregate/outcome/receipts/outgoing intent, then ACKs the command.

The event reaction owns hand-off recovery. The command consumer owns subsequent
business execution, retries and dead-letter recovery. A handler never directly
calls another command handler.

## Python: pair the consumers

### Python

Source: [preparation.py](../../../services/operations/src/operations/apps/composition/preparation.py).

```python source="../../../services/operations/src/operations/apps/composition/preparation.py"
container.commands().subscription(
    container.codec(),
    definitions,
    ORDER_PLACED,
    lambda event: derived_id("ticket", event["orderId"]),
    container.accepted().handle,
    AcceptOrderCommandHandler,
)
```

The definition supplies OrderPlaced's typed incoming contract. The target ticket
identity is derived from the order, and `AcceptOrderCommandHandler` receives a
fresh write repository when the queued command eventually executes.

## TypeScript: retain business identity

### TypeScript

Source: [loyalty.ts](../../../services/engagement/src/apps/composition/loyalty.ts).

```ts source="../../../services/engagement/src/apps/composition/loyalty.ts"
    ...c.accountCommands.subscription(
      CollectionIntegrationEvent.ORDER_COLLECTED,
      c.creditCodec,
      definitions,
      (event) => event.customerId,
      c.collected,
      (repository, realtime) => new CreditCollectionCommandHandler(repository, realtime, derivedId),
      (m, command) => ({ ...m, id: derivedId('collection-credit', command.orderId) }),
    ),
```

The metadata mapping derives the credit command identity from `orderId`, while
the target account comes from `customerId`. This protects business credit across
distinct event IDs representing the same collection.

## Go: record collection on the Order

The running chain is `OrderCollectedIntegrationEvent` →
`OrderCollectedIntegrationEventHandler` → `CommandBusPort.Enqueue` → private
`RecordCollectionCommand` → `RecordCollectionCommandHandler` → `Order.RecordCollection`.
Collection owns physical handover. Ordering records that fact on its existing root.
A completed Pickup with `Order.collectionRecorded == false` is an allowed pending
state while the private command awaits execution.

### 1. Name the invariant and handler behaviour

After generating the contract in step 2, follow [record_collection.go](../../../services/storefront/contexts/ordering/application/commands/record_collection.go).
Inject `OrderWriteRepository` and `RealtimeBusPort`, load the target, call
`Order.RecordCollection(customerID)`, save a changed root and explicitly publish
its snapshot. The aggregate requires a placed order and matching customer.
Repeated collection facts produce no new revision, fact or browser publication.
Missing order authority is retryable; a mismatched customer is a recorded rejection.

```go source="../../../services/storefront/contexts/ordering/application/commands/record_collection.go"
func (h RecordCollectionCommandHandler) Execute(ctx context.Context, m a.CommandContext, c orderingcontracts.RecordCollectionCommand) (a.CommandResult, error) {
	loaded, err := h.Repository.Get(ctx, m.Target)
	if err != nil {
		return a.CommandResult{}, err
	}
	if !loaded.Exists {
		return a.CommandResult{}, fmt.Errorf("collected order is missing")
	}
	order := loaded.State
	if err = order.RecordCollection(c.CustomerID); err != nil {
		return a.CommandResult{}, err
	}
	if len(order.Events()) > 0 {
		if err = h.Repository.Save(ctx, order); err != nil {
			return a.CommandResult{}, err
		}
		if err = h.Realtime.Publish(ctx, a.RealtimeEvent{Snapshot: order.Snapshot()}); err != nil {
			return a.CommandResult{}, err
		}
	}
	return a.Result("collection_recorded"), nil
}
```

### 2. Define and preserve the private wire format

Add the payload and `CommandEnvelope.command` oneof case in
[internal_commands.proto](../../../services/storefront/contexts/ordering/adaptors/messaging/internal_commands.proto).
Declare `owner: "ordering"` and `visibility: PRIVATE`, then run `pnpm generate:contracts`.
Import `RecordCollectionCommand` from the [generated plain contracts](../../../services/storefront/contracts/requests/generated/model/ordering/payloads.go).
Use a new field number. The shared envelope fields retain the source receipt and
command identity. This format belongs to Ordering's adaptor and is separate from
the API's published request envelope.

```protobuf source="../../../services/storefront/contexts/ordering/adaptors/messaging/internal_commands.proto"
message CommandEnvelope {
 string id = 1;
 string consumer = 2;
 string source_event_id = 3;
 string source_hash = 4;
 string target = 5;
 string correlation_id = 6;
 bytes receipt_material = 7;
 oneof command { RecordCollectionCommand record_collection = 10; }
}
message RecordCollectionCommand {
  option (cafe.requests.v1.contract) = { owner: "ordering" visibility: PRIVATE response: "cafe.requests.v1.Outcome" };
  string customer_id = 1; }
```

Implement `Encode` and `Decode` in
[RecordCollectionCodec](../../../services/storefront/contexts/ordering/adaptors/messaging/record_collection_codec.go).
They call generated payload mappings and validate UUIDs, subscription ownership,
the source hash and original receipt JSON. `ID` must equal
`DerivedID(consumer, sourceEventID)`; `CausationID` retains that source event.
Keep a historical hex fixture and native round-trip/invalid-envelope tests.
A later mapping must preserve command bytes already accepted into the outbox.

### 3. Declare the source and generated identifiers

Add `collection` to Storefront's `events.consumes` in
[generation.json](../../../contracts/generation.json). Declare the consumer in
[subscriptions.json](../../../services/storefront/contexts/ordering/adaptors/messaging/subscriptions.json):

```json
{
  "consumer": "ordering.record-collection",
  "event": "collection.order-collected",
  "command": "recordCollection"
}
```

Run `pnpm generate:contracts`. `recordCollection` is the Protobuf oneof field's
JSON name. Generation validates it and produces
`OrderingPrivateCommandRecordCollection`, `OrderingSubscriptionRecordCollection`,
the Collection DTO/codec, and event/command queue permissions. The handler type
is selected explicitly in composition; the generated string alone cannot select it.

### 4. Map the fact and inject the durable bus

```go source="../../../services/storefront/contexts/ordering/application/eventhandlers/order_collected.go"
// OrderCollectedIntegrationEventHandler durably requests an owner transition before acknowledging the fact.
type OrderCollectedIntegrationEventHandler struct {
	Commands a.CommandBusPort[orderingcontracts.RecordCollectionCommand]
}

func (h OrderCollectedIntegrationEventHandler) Handle(ctx context.Context, m a.Metadata, event model.OrderCollectedIntegrationEvent) (a.Outcome, error) {
	return h.Commands.Enqueue(ctx, m, orderingcontracts.RecordCollectionCommand{CustomerID: event.CustomerID})
}
```

The Fx provider in [ordering.go](../../../services/storefront/apps/storefront/di/ordering.go)
injects `store.NewCommandOutbox[orderingcontracts.RecordCollectionCommand](database, codec)`.
The target is `event.OrderID`; the source envelope's aggregate is a Pickup, so
using its aggregate ID would target the wrong root.

### 5. Register the codec and plain handler together

```go source="../../../services/storefront/apps/storefront/di/ordering.go"
registration.Subscribe(requests.Commands, s, rpc.Subscription(names.OrderingSubscriptionRecordCollection), commandCodec,
		func(event model.OrderCollectedIntegrationEvent) string { return event.OrderID }, collected.Handle,
		func(repository ports.OrderWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) commands.RecordCollectionCommandHandler {
			return commands.RecordCollectionCommandHandler{Repository: repository, Realtime: realtime}
		})
```

The aggregate registry supplies `NewOrderTransaction` once. `Subscribe` constructs
the central executor and registers both `ordering.record-collection` and
`ordering.record-collection.command`, plus the relay decoder. Startup rejects a
missing consumer, mismatched binding or absent command decoder. Every invocation
receives a scoped repository and realtime buffer; application code has no SQL or
container dependency.

### 6. Apply the migration and verify recovery

Run `pnpm dev:up` after generation. The administrator applies the append-only
`0004_internal_commands.sql` migrations and generated broker permissions.
Runtime credentials can insert immutable acceptance records and update dispatch
progress; they cannot replace accepted bytes.

Run `pnpm test:focused storefront --context ordering`, `pnpm verify` and
`pnpm test:integration`. The real infrastructure lane pauses the command consumer,
observes committed acceptance and publication, restarts Storefront and verifies the
Order transition. PostgreSQL tests cover rollback, duplicate facts and fenced
leases; RabbitMQ tests cover bounded retries, dead letters and explicit replay.
Use `pnpm workflow:inspect <correlation-uuid>` to distinguish accepted, published
and executed work. After repairing a failed dependency, explicitly replay with
`pnpm events:replay ordering.record-collection.command`.

## Durable evidence

The Python outbox stores the encoded body before returning acceptance:

### Python

Source: [command_outbox.py](../../../services/operations/src/operations/adaptors/persistence/command_outbox.py).

```python source="../../../services/operations/src/operations/adaptors/persistence/command_outbox.py"
connection.execute(
    """INSERT INTO cafe.internal_commands(id,consumer,event_id,fingerprint,target,body)
    VALUES(%s,%s,%s,%s,%s,%s)""",
    (
        metadata.id,
        metadata.consumer,
        metadata.source_id,
        metadata.source_hash,
        metadata.target,
        body,
    ),
)
connection.execute(
    "INSERT INTO cafe.internal_command_dispatches(event_id) VALUES(%s)", (metadata.id,)
)
```

The TypeScript [command outbox](../../../services/engagement/src/adaptors/persistence/command-outbox.ts)
has the same receiving-evidence and dispatch boundary. Private codecs and stored
compatibility fixtures live with the owning context's messaging adaptors.

## Related concepts

[Commands](commands.md) · [Integration event handlers](integration-event-handlers.md) · [Domain event handlers](domain-event-handlers.md) · [Outbox](outbox.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
