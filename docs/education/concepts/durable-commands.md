# Durable commands

A durable command is an owner command whose exact queued representation and
dispatch intent survive a process failure. Event reactions use it to transfer
responsibility before acknowledging the incoming event.

## Creation recipe

1. Define the target owner command/handler and its invariant, aggregate and permitted pending state.
2. Add the private command payload/envelope in the owner's `internal_commands.proto`, then implement its typed codec and semantic validation with compatibility fixtures.
3. Inject `CommandBusPort` into the event reaction and explicitly map the fact into the owner DTO. Choose stable target/command identities; correlation groups a workflow and does not deduplicate it.
4. Declare the source event, stable subscription identity and private command in `subscriptions.json`. Run `pnpm generate:contracts` for generated identifiers and topology.
5. Bind the concrete durable command outbox, reaction and plain command handler through the shared paired-consumer helper. Preserve separate event-acceptance and command-execution transactions.
6. Test acceptance rollback, repeated delivery, lost confirms/ACKs, retained command bytes and replay. Python/TypeScript have concrete examples below; Go currently requires the missing private hand-off implementation described in [coverage](../implementation-coverage.md).

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

## Go's current scope

Storefront's existing event reactions update projections. It has no corresponding
aggregate-changing private command subscription example. Its [published commands](commands.md)
still use the central executor and durable request/reply semantics. A new private
command reaction needs the explicit infrastructure described in
[decision 0011](../../decisions/0011-durable-commands-before-aggregate-mutations.md).

### Go: illustrative hand-off shape

This is the application shape a Go equivalent of the Preparation reaction would
need. It is a teaching example, with imports omitted. The port, private codec,
outbox implementation and consumer pair are absent from Storefront today.
`a.Metadata`, `a.Outcome` and `model.OrderPlacedIntegrationEvent` refer to the
existing [application metadata](../../../services/storefront/foundation/application/command.go)
and [generated event DTO](../../../services/storefront/contracts/events/model/ordering_generated.go).

```go
type AcceptOrderCommand struct {
	OrderID    string
	CustomerID string
}

type CommandBusPort interface {
	Enqueue(context.Context, a.Metadata, AcceptOrderCommand) (a.Outcome, error)
}

type OrderPlacedReaction struct {
	Commands CommandBusPort
}

func (h OrderPlacedReaction) Handle(ctx context.Context, m a.Metadata, event model.OrderPlacedIntegrationEvent) (a.Outcome, error) {
	return h.Commands.Enqueue(ctx, m, AcceptOrderCommand{
		OrderID: event.OrderID, CustomerID: event.CustomerID,
	})
}
```

The shortened command demonstrates the hand-off; a real owner contract would
also include the preparation instructions and its compatibility fixtures. An
in-memory queue cannot fulfil this port's required atomic receiving-evidence and
dispatch guarantees. The [coverage report](../implementation-coverage.md) records
the missing Go infrastructure.

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
