# Commands

A command requests an owner to perform a business action. Creating its application
DTO is one step in making that action callable. This recipe connects a published
Protobuf request to a plain command handler in Go, Python and TypeScript.

Follow the steps for your owning language. The examples already exist in this
checkout: use their files as templates for a new action, and allocate new names
and unused Protobuf field numbers. Reuse the existing aggregate infrastructure.
For an event-triggered private command, follow the [durable command recipe](../../developer-workflow.md#wire-an-aggregate-changing-subscription)
after defining the application command and handler.

## 1. Choose the owner and transition

Run `pnpm context ordering`, `pnpm context preparation` or `pnpm context loyalty`
and read the linked context README. Write down the target aggregate, input,
required starting state, business rejection and resulting state.

| Example             | Owner / language     | Aggregate         | Transition                                                  |
| ------------------- | -------------------- | ----------------- | ----------------------------------------------------------- |
| PlaceOrder          | Ordering / Go        | Order             | Valid non-empty draft → placed, with OrderPlacedDomainEvent |
| CompletePreparation | Preparation / Python | PreparationTicket | Preparing → ready, with DrinksReadyDomainEvent              |
| RedeemReward        | Loyalty / TypeScript | Reward            | Issued and unexpired → redeemed for an order                |

These are immediate transitions of existing roots. Downstream integration work
can remain pending after commit. A new aggregate also needs its own restoration,
repository and transaction bindings; follow [write repositories](write-repositories.md)
and [dependency injection](dependency-injection.md) before registering its first command.

## 2. Write the application command and handler

Create one command DTO/handler file under the context's `application/commands/`:
use snake_case in Go/Python and kebab-case in TypeScript. Put business input in
the DTO. Target identity, command identity, expected revision and correlation
metadata arrive separately through the executor. These are the actual DTOs:

### Go

Source: [place_order.go](../../../services/storefront/contexts/ordering/application/commands/place_order.go).

```go source="../../../services/storefront/contexts/ordering/application/commands/place_order.go"
type PlaceOrderCommand struct{}
```

### Python

Source: [complete_preparation.py](../../../services/operations/src/operations/contexts/preparation/application/commands/complete_preparation.py).

```python source="../../../services/operations/src/operations/contexts/preparation/application/commands/complete_preparation.py"
class CompletePreparationCommand(TypedDict):
    """Requests completion of the targeted preparation ticket."""

    pass
```

### TypeScript

Source: [redeem-reward.ts](../../../services/engagement/src/contexts/loyalty/application/commands/redeem-reward.ts).

```ts source="../../../services/engagement/src/contexts/loyalty/application/commands/redeem-reward.ts"
export type RedeemRewardCommand = { orderId: string };
```

Add the named handler in that same file. Inject `OrderWriteRepository`,
`TicketWriteRepository` or `RewardWriteRepository`; load the target, handle missing
state, invoke the aggregate operation and save. The [command handler page](command-handlers.md#the-complete-use-cases-in-three-languages)
contains the complete files for all three examples, including imports and method
signatures. Implement and test the corresponding aggregate behaviour before wiring
transport. The central executor will own the transaction and repository lifetime.

**Checkpoint:** the command DTO and handler can be tested using plain ports.
They have no imports from Protobuf, RabbitMQ or the DI framework.

## 3. Declare the published payload and envelope case

Edit two owner schema files: the payload under `messaging/commands/v1/` and its
case inside `message Command` → `oneof payload` under `messaging/v1/`.
A payload declaration alone does not create a routing identifier.

### Go: PlaceOrder

Payload: [ordering_commands.proto](../../../contracts/ordering/messaging/commands/v1/ordering_commands.proto).

```protobuf source="../../../contracts/ordering/messaging/commands/v1/ordering_commands.proto"
message PlaceOrder {
}
```

Envelope: [ordering_requests.proto](../../../contracts/ordering/messaging/v1/ordering_requests.proto), inside `Command.payload`.

```protobuf source="../../../contracts/ordering/messaging/v1/ordering_requests.proto"
cafe.ordering.requests.v1.PlaceOrder place_order = 20;
```

### Python: CompletePreparation

Payload: [preparation_commands.proto](../../../contracts/preparation/messaging/commands/v1/preparation_commands.proto).

```protobuf source="../../../contracts/preparation/messaging/commands/v1/preparation_commands.proto"
message CompletePreparation {
}
```

Envelope: [preparation_requests.proto](../../../contracts/preparation/messaging/v1/preparation_requests.proto), inside `Command.payload`.

```protobuf source="../../../contracts/preparation/messaging/v1/preparation_requests.proto"
cafe.preparation.requests.v1.CompletePreparation complete_preparation = 22;
```

### TypeScript: RedeemReward

Payload: [loyalty_commands.proto](../../../contracts/loyalty/messaging/commands/v1/loyalty_commands.proto).

```protobuf source="../../../contracts/loyalty/messaging/commands/v1/loyalty_commands.proto"
message RedeemReward {
  optional string order_id = 1 [(cafe.requests.v1.required_input) = true];
}
```

Envelope: [loyalty_requests.proto](../../../contracts/loyalty/messaging/v1/loyalty_requests.proto), inside `Command.payload`.

```protobuf source="../../../contracts/loyalty/messaging/v1/loyalty_requests.proto"
cafe.loyalty.requests.v1.RedeemReward redeem_reward = 24;
```

For required scalar input, follow `order_id`: explicit presence plus the
`required_input` annotation, with the validation schema imported. Preserve existing
field numbers, names and compatibility. The numbers above belong to the existing
examples; a new command gets a fresh case number.

## 4. Generate the wire types and routing identifiers

From the repository root, run:

```sh
pnpm generate:contracts
```

The [identifier generator](../../../scripts/generate_request_names.py) reads the
compiled `Command.payload` fields. It derives member names from the context and
snake_case field name, and routing values from the field's Protobuf JSON name.
By default, `place_order` has the JSON name `placeOrder`.
It does not inspect application command classes.

| Envelope case                      | Generated identifier                      | Routing value         |
| ---------------------------------- | ----------------------------------------- | --------------------- |
| Ordering `place_order`             | `OrderingCommandPlaceOrder`               | `placeOrder`          |
| Preparation `complete_preparation` | `PreparationCommand.COMPLETE_PREPARATION` | `completePreparation` |
| Loyalty `redeem_reward`            | `LoyaltyCommand.REDEEM_REWARD`            | `redeemReward`        |

### Go output

[names.go](../../../services/storefront/contracts/requests/generated/cafe/requests/v1/names.go)
contains a typed string constant. Its package is imported as `names` in composition.

```go source="../../../services/storefront/contracts/requests/generated/cafe/requests/v1/names.go"
OrderingCommandPlaceOrder     OrderingCommand = "placeOrder"
```

### Python output

[request_names.py](../../../services/operations/src/operations/adaptors/generated/request_names.py)
contains a `StrEnum` of routing names. `PreparationCommand` is that enum;
`CompletePreparationCommand` is the application DTO written in step 2.

```python source="../../../services/operations/src/operations/adaptors/generated/request_names.py"
class PreparationCommand(StrEnum):
    START_PREPARATION = 'startPreparation'
    COMPLETE_PREPARATION = 'completePreparation'
```

### TypeScript output

[request-names.ts](../../../services/engagement/src/adaptors/generated/request-names.ts)
contains the corresponding string enum.

```ts source="../../../services/engagement/src/adaptors/generated/request-names.ts"
export enum LoyaltyCommand {
  REDEEM_REWARD = 'redeemReward',
}
```

**Checkpoint:** inspect the generated diff for the payload binding, envelope case
and identifier. Commit generated outputs with their schema sources. Change schema
sources and rerun generation whenever the published contract changes.

## 5. Write the wire-to-application parser

Create or extend the owning context's messaging adaptor. Import the generated
payload type and the application DTO, validate boundary input, and return the
application command. This is handwritten conversion code.

### Go

[ordering_requests.go](../../../services/storefront/contexts/ordering/adaptors/messaging/ordering_requests.go)
uses `pb` for the generated Ordering payload package and `commands` for the
application command package:

```go source="../../../services/storefront/contexts/ordering/adaptors/messaging/ordering_requests.go"
import (
	commands "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contexts/ordering/application/commands"
	pb "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/requests/generated/cafe/requests/v1/contexts/ordering"
)
```

```go source="../../../services/storefront/contexts/ordering/adaptors/messaging/ordering_requests.go"
func PlaceOrder(p *pb.PlaceOrder) commands.PlaceOrderCommand {
	return commands.PlaceOrderCommand{}
}
```

### Python

[ticket_requests.py](../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py)
imports `CompletePreparation` from the generated `preparation_commands_pb2` module
and `CompletePreparationCommand` from the application module linked in step 2:

```python source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py"
from google.protobuf.message import Message
```

```python source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py"
from operations.contexts.preparation.application.commands.complete_preparation import (
    CompletePreparationCommand,
)
from operations.adaptors.generated.cafe.requests.v1.contexts.preparation.preparation_commands_pb2 import (
    StartPreparation,
    CompletePreparation,
)
```

The existing adaptor also handles StartPreparation. Keep the imports needed by
the parsers in your file, then add the conversion function:

```python source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/ticket_requests.py"
def complete(value: Message) -> CompletePreparationCommand:
    if not isinstance(value, CompletePreparation):
        raise ValueError("Unexpected complete preparation payload")
    return CompletePreparationCommand()
```

`CompletePreparationCommand()` creates an empty dictionary with a TypedDict shape.
The parser's return annotation carries the application type to the registration.

### TypeScript

[reward-requests.ts](../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-requests.ts)
uses generated wire types and owner validation helpers:

```ts source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-requests.ts"
import { identifier } from '../../../../foundation/domain.js';
```

```ts source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-requests.ts"
import { record, text } from '../../../../adaptors/messaging/request-values.js';
import type { loyalty as wire } from '../../../../adaptors/generated/request-types.js';
```

```ts source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/reward-requests.ts"
export function redeem(value: unknown): { orderId: string } {
  const fields = record(value);
  const wire: wire.RedeemReward = { orderId: text(fields.orderId) };
  return { orderId: identifier(wire.orderId!) };
}
```

The returned `{ orderId: string }` is structurally compatible with
`RedeemRewardCommand`. The current parser writes that shape inline; it does not
import the type alias. Its `wire` type import is the generated Loyalty namespace,
while composition's `wire` below is an alias for this handwritten parser module.

**Checkpoint:** test valid input, missing required fields and invalid values.
The parser produces application input; the aggregate checks lifecycle rules.

## 6. Bind the parser and handler in composition

Use the owning aggregate's existing command registration and transaction. Import
the generated identifier, the parser module and the concrete handler. The aliases
below are local import names chosen in these files:

| Composition name            | Meaning and import target                                                                                            |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Go `names`                  | Generated [request identifiers](../../../services/storefront/contracts/requests/generated/cafe/requests/v1/names.go) |
| Go `rpc`                    | Handwritten [Ordering messaging package](../../../services/storefront/contexts/ordering/adaptors/messaging/)         |
| Go `orderingcommands`       | Plain [application command package](../../../services/storefront/contexts/ordering/application/commands/)            |
| Python `PreparationCommand` | Enum from `operations.adaptors.generated.request_names`                                                              |
| Python `wire`               | `ticket_requests` imported from `operations.contexts.preparation.adaptors.messaging`                                 |
| TypeScript `LoyaltyCommand` | Enum imported from `../../adaptors/generated/request-names.js`                                                       |
| TypeScript `wire`           | `import * as wire from '../../contexts/loyalty/adaptors/messaging/reward-requests.js'`                               |

In particular, `rpc.PlaceOrder` is the handwritten function from step 5.
`pb.PlaceOrder` inside that function is the generated Protobuf message type.
Go's import alias places the function under `rpc` at this call site.

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

These statements belong inside the existing request-registration functions.
Go's `commands` is created with `registration.New(r, transaction)`; Python's
`commands` is resolved from `container.commands()`; TypeScript's `rewardCommands`
is constructed with `new CommandRegistration(c.rewardTransaction)`. See the
[transaction bindings](dependency-injection.md#configure-one-aggregate-transaction).
For a command on an existing root, reuse that registration. Resolve any additional
handler dependency at startup and capture it in the factory, as the TypeScript
example does with `clock`.

The three roles in each registration are **routing identifier → parser → handler
factory**. At runtime the registry selects the identifier and calls the parser;
the executor opens the transaction and supplies the factory with its repository
and invocation-scoped realtime bus. The constructed handler receives the parser's result in `execute` / `Execute`.

| Language   | Parser result passed to the handler                                                                          |
| ---------- | ------------------------------------------------------------------------------------------------------------ |
| Go         | `rpc.PlaceOrder` returns `commands.PlaceOrderCommand`; `PlaceOrderCommandHandler.Execute` accepts it         |
| Python     | `wire.complete` returns `CompletePreparationCommand`; `CompletePreparationCommandHandler.execute` accepts it |
| TypeScript | `wire.redeem` returns the `RedeemRewardCommand` shape; `RedeemRewardCommandHandler.execute` accepts it       |

The application DTO is therefore connected through the parser output and handler
input types. Each registration supplies an explicit factory; the Python lambda
constructs its handler with only the repository because this use case delegates
realtime publication to a domain-event subscriber. Registration resolves the
handler from this declaration. Static checks cover parser/handler compatibility; a dispatch test must
also prove that the selected wire case reaches the intended behaviour.

## 7. Expose an HTTP operation when required

For an API-callable command, add the operation and request schema to the owner's
HTTP contract: [Ordering](../../../contracts/ordering/http_api/orders.paths.json),
[Preparation](../../../contracts/preparation/http_api/tickets.paths.json) or
[Loyalty](../../../contracts/loyalty/http_api/rewards.paths.json), plus that directory's
`schemas.json`. Preserve the existing command metadata/reply conventions.

Add its operation ID, Protobuf message and payload field mapping to
[contract_mappings.json](../../../services/api/adaptors/http/backend/contract_mappings.json).
The existing `redeemReward` entry maps HTTP `orderId` to Protobuf `orderId`;
`placeOrder` and `completePreparation` have empty payload field maps.
Run `pnpm generate:contracts` again to update HTTP, API mapping and client outputs.
Implement any new translation behaviour in the API's [HTTP backend](../../../services/api/adaptors/http/backend/).
Owner services expose commands over messaging; business HTTP routes belong in the API.

A browser feature then calls the generated operation client and uses the existing
[command recovery and realtime flow](../../../CLIENT-SUBSCRIPTIONS.md). A new action
on an existing aggregate can reuse its current projection contract. If the view
changes, update that contract and its mappings as part of the feature.

## 8. Prove the whole connection

Add domain/handler tests for success, invalid state and missing targets. Extend a
production-composition dispatch test using a generated request and a test
transaction. Assert the parsed input, selected aggregate, saved state, outcome,
metadata and expected private facts. Test malformed input and expected-version
conflicts. Use these existing tests as templates:

- Go: [menu_requests_test.go](../../../services/storefront/apps/storefront/di/menu_requests_test.go).
- Python: [test_command_registration.py](../../../services/operations/tests/test_command_registration.py).
- TypeScript: [command-registration.test.ts](../../../services/engagement/src/apps/command-registration.test.ts).

Add the business scenario/native binding and update the context's navigation and
catalogue where needed. For an HTTP operation, extend API translation and live
operation tests. Run the appropriate owner lane, then the repository gates:

```sh
pnpm test:focused storefront --context ordering
# Or: pnpm test:focused operations --context preparation
# Or: pnpm test:focused engagement --context loyalty
pnpm check --changed
pnpm verify
pnpm test:integration
```

The recipe is complete when a real boundary request reaches the concrete handler,
commits the expected state and returns its recorded outcome. An identical retry
retains the same command identity, input and expected revision. Infrastructure
checks establish rollback and lost-reply recovery; record executed evidence in
[verification](../../verification.md).

## Related concepts

[Command handlers](command-handlers.md) · [Application layer](application-layer.md) · [Contracts and generation](contracts-and-generation.md) · [Durable commands](durable-commands.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
