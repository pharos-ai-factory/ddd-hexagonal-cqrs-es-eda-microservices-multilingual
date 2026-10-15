# Command and query responses

The response is part of the versioned request contract. Protobuf owns its fields,
nested resources, optional presence and request association. Generation supplies
plain DTOs and boundary encoders in Go, Python and TypeScript.

## Creation recipe

1. Define the success response in the owning reply schema. Commands use the shared
   `Outcome`; queries select their owner resource or collection response.
2. Set the request's `contract.response` to the fully qualified response message.
   Public responses must have one matching success member in the owner `Reply` envelope.
3. Define nested/repeated fields there. Mark required message values with
   `required_output`; retain `optional` for fields whose absence matters.
4. Run `pnpm generate:contracts`. Import the generated plain types. Read-model
   aliases can retain the port's vocabulary without redeclaring any fields.
5. For commands, return the internal business result from the handler. The executor
   assembles the generated `Outcome` with the committed revision or rejection and
   stages exact reply bytes in the same transaction.
6. For queries, return the generated query response. Select success or `RequestError`
   explicitly and register through the generated owner binding.
7. Test response semantics, wire encoding and compatibility. A schema change must
   update its plain DTO, encoder and typed callers through generation.

## Command outcome types in all three languages

The shared [common.proto](../../../contracts/shared/messaging/v1/common.proto)
defines `Outcome`, `Rejection` and `RequestError`. Each command response is a
generated alias of `Outcome`, including private commands whose outcome is retained
for delivery and retries. The feature handler's `CommandResult` carries the status
chosen by its business operation; the executor supplies the final revision.

### Go

```go source="../../../services/storefront/contracts/requests/generated/model/shared/responses.go"
type Outcome struct {
	AggregateID string     `json:"aggregateId"`
	Version     uint64     `json:"version"`
	Status      string     `json:"status"`
	Rejection   *Rejection `json:"rejection,omitempty"`
}
```

### Python

```python source="../../../services/operations/src/operations/contracts/requests/generated/shared_responses.py"
class Outcome(TypedDict):
    """Schema-generated plain response payload."""
    aggregateId: str
    version: int
    status: str
    rejection: NotRequired[Rejection]
```

### TypeScript

```ts source="../../../services/engagement/src/contracts/requests/generated/shared-responses.ts"
export type Outcome = {
  aggregateId: string;
  version: number;
  status: string;
  rejection?: Rejection;
};
```

The encoders preserve the schema's scalar bounds and optional rejection. Expected
business errors become the published code/message DTO. Domain errors keep their
own behaviour and dependencies. Invalid output aborts encoding; an API command
cannot commit successfully when its exact reply bytes cannot be produced.

## Query responses and their bindings

The [query recipe](queries.md) contains complete handlers and composition examples
for `GetOrder`, `GetTicket` and `GetReward`. Their generated response types constrain
the successful payload and the shared error alternative. A handler for a different
response cannot satisfy the generated registration signature. The encoder rejects
an ambiguous response containing both alternatives.

Application handlers choose missing-resource behaviour and pagination mode.
Persistence adaptors validate stored authority and select response fields. Generated
encoders copy those DTO fields into wire messages. All current published read views
are aliases or re-exports of generated resource DTOs. Generic repository containers
and aggregate snapshots retain their internal roles.

## Change a field safely

Change the response Protobuf first and regenerate. Update the persistence selection
or handler construction wherever a new required field is needed. Optional additions
preserve absence; wire field numbers and historical fixtures remain stable.
Generated contract drift, source-linked documentation excerpts and native type checks
must pass. Then run the infrastructure lane to prove exact command reply recovery
and query behaviour through the API and browser.

The supported response vocabulary includes strings, booleans, signed/unsigned
32-bit and 64-bit integers, nested messages, repeated fields and optional presence. Unsupported maps, enums, recursive messages and user
oneofs fail generation until their semantics and fixtures are implemented.

[Commands](commands.md) · [Queries](queries.md) · [Read models](read-models.md) ·
[Transactions](transactions.md) · [Contracts and generation](contracts-and-generation.md)
