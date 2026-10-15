# Application layer

The application layer exposes the operations a bounded context offers and the
capabilities those operations require. It coordinates work around the domain
model and gives callers explicit input and result types.

## Responsibilities

| Role               | Coordinates              | Delegates to                                     |
| ------------------ | ------------------------ | ------------------------------------------------ |
| Command handler    | One requested transition | Aggregate and named write repository             |
| Query handler      | One question             | Named read repository and application read model |
| Event handler      | A delivered fact         | Durable command or publication port              |
| Projection handler | A derived view update    | Projection persistence port                      |

Application code stays independent of transports, databases and DI frameworks.
It may depend on its own domain and published contract DTOs at the documented
event boundary. It never imports another context's implementation.

## Concrete use cases

The entry point signatures show the same responsibility in each language:

### Go

Source: [place_order.go](../../../services/storefront/contexts/ordering/application/commands/place_order.go).

```go source="../../../services/storefront/contexts/ordering/application/commands/place_order.go"
type PlaceOrderCommandHandler struct {
	Repository ports.OrderWriteRepository
}

func (h PlaceOrderCommandHandler) Execute(ctx context.Context, m a.CommandContext, _ orderingcontracts.PlaceOrderCommand) (a.CommandResult, error) {
```

### Python

Source: [complete_preparation.py](../../../services/operations/src/operations/contexts/preparation/application/commands/complete_preparation.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/commands/complete_preparation.py"
class CompletePreparationCommandHandler:
    """Completes one ticket and records the outgoing preparation fact atomically."""

    def __init__(self, repository: TicketWriteRepository) -> None:
        self.repository = repository

    def execute(self, context: CommandContext, command: CompletePreparationCommand) -> CommandResult:
```

### TypeScript

Source: [redeem-reward.ts](../../../services/engagement/src/contexts/loyalty/application/commands/redeem-reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/commands/redeem-reward.ts"
export class RedeemRewardCommandHandler {
  constructor(
    private repository: RewardWriteRepository,
    private realtime: RealtimeBusPort<RealtimeEvent>,
    private clock: () => Date,
  ) {}
  async execute(context: CommandContext, command: RedeemRewardCommand) {
```

These signature excerpts continue in [command handlers](command-handlers.md),
which shows their complete bodies and constructor dependencies.

## Keep decisions with their owner

The handler can decide that missing state is a use-case rejection. Whether an
existing Order may be placed belongs to the domain. The framework determines
how to decode a message; persistence determines how to restore state. These
responsibilities meet through plain [ports](ports.md).

Command and query inputs are generated from their authoritative Protobuf schemas.
Their Go structs, Python TypedDicts/dataclasses and TypeScript types have no HTTP
or Protobuf runtime dependency. Handlers depend on the contract's meaning and
shape; generated [adaptor](adaptors.md) mappings handle its wire representation.
Owner adaptors retain semantic validation.

## Related concepts

[Commands](commands.md) · [Command handlers](command-handlers.md) · [Query handlers](query-handlers.md) · [Domain layer](domain-layer.md) · [Ports](ports.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
