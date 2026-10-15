# Dependency injection and composition roots

Dependency injection supplies a component's collaborators from outside. The
composition root selects concrete implementations and their lifetime. This
repository uses Fx in Go, Dependency Injector in Python and Awilix in TypeScript.

## Configure one aggregate transaction

These composition excerpts connect the context database, owner restoration,
write repository and private-event selection. The shared command registration
then uses this transaction for every command on that aggregate.

### Go

Source: [ordering.go](../../../services/storefront/apps/storefront/di/ordering.go).

```go source="../../../services/storefront/apps/storefront/di/ordering.go"
func(s *support.StorefrontRuntime) execution.Transaction[*d.Order] {
	return pg.NewOrderTransaction(s.Databases["ordering"])
},
```

### Python

Source: [preparation.py](../../../services/operations/src/operations/apps/composition/preparation.py).

```python source="../../../services/operations/src/operations/apps/composition/preparation.py"
ticket_transaction = providers.Singleton(
    PostgresAggregateTransaction[TicketSnapshot, PreparationTicket],
    database,
    "ticket",
    restore_ticket_snapshot,
    PostgresTicketWriteRepository,
    ticket_publications,
)
```

### TypeScript

Source: [loyalty.ts](../../../services/engagement/src/apps/composition/loyalty.ts).

```ts source="../../../services/engagement/src/apps/composition/loyalty.ts"
rewardTransaction: asFunction(
  (c: LoyaltyDependencies) =>
    new PostgresAggregateTransaction(
      c.database,
      'reward',
      restoreRewardSnapshot,
      (source) => new PostgresRewardWriteRepository(source),
      rewardPublications,
    ),
).singleton(),
```

Go's `pg.NewOrderTransaction` binds the owner repository and event selector in
its own context adaptor. Python and TypeScript declare the same dependencies
directly in their transaction provider.

## Construct the handler inside the invocation

First complete the [command creation recipe](commands.md): define the application
DTO and handler, add the published schema/envelope case, generate the identifier
and write the owner parser. The registration below connects those completed parts.

### Go

Source: [ordering_requests.go](../../../services/storefront/apps/storefront/di/ordering_requests.go).

```go source="../../../services/storefront/apps/storefront/di/ordering_requests.go"
	registration.Register(commands, names.OrderingCommandPlaceOrder, rpc.PlaceOrder, func(repository ports.OrderWriteRepository, realtime a.RealtimeBusPort[a.RealtimeEvent]) orderingcommands.PlaceOrderCommandHandler {
		return orderingcommands.PlaceOrderCommandHandler{Repository: repository}
	})
```

### Python

Source: [preparation.py](../../../services/operations/src/operations/apps/composition/preparation.py).

```python source="../../../services/operations/src/operations/apps/composition/preparation.py"
    commands.request(
        registry,
        PreparationCommand.COMPLETE_PREPARATION,
        wire.complete,
        lambda repository, _realtime: CompletePreparationCommandHandler(repository),
    )
```

### TypeScript

Source: [loyalty.ts](../../../services/engagement/src/apps/composition/loyalty.ts).

```ts source="../../../services/engagement/src/apps/composition/loyalty.ts"
c.rewardCommands.request(
  registry,
  LoyaltyCommand.REDEEM_REWARD,
  wire.redeem,
  (repository, realtime) => new RedeemRewardCommandHandler(repository, realtime, clock),
);
```

The registration helper creates the executor. Its factory receives the fresh
repository after the transaction boundary is established. This lifetime prevents
a handler from retaining a repository belonging to an earlier request.

Query handlers have a different lifetime: they receive read repositories during
composition and may be reused because those repositories acquire a connection
for each read. Read the [query handler](query-handlers.md) examples alongside this page.

## Resolve and dispose deliberately

Resolve the graph before workers start. Startup failures clean up resources
already created; shutdown drains in-flight work before disposing pools and
containers. Constructors in the application and domain receive plain collaborators.
Container resolution and framework imports stay under `apps/`.

Dependency inversion describes the direction of source dependencies towards
policy and its ports. Dependency injection is the mechanism used to supply an
implementation. Constructor injection can establish that mechanism without a container.

## Related concepts

[Command registration](commands.md) · [Ports](ports.md) · [Adaptors](adaptors.md) · [Transactions](transactions.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
