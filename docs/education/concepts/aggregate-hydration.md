# Aggregate hydration

Hydration restores an aggregate from its persisted representation. This reference
uses current-state persistence: the stored snapshot is authoritative, and the
owner's restoration path validates it before exposing domain behaviour.

## Creation recipe

1. Define the persisted snapshot shape and immutable root identity in the owner context.
2. Implement a restore function that validates every stored invariant and constructs the domain object with no newly pending events. Corrupt storage is a technical failure.
3. Connect restore and snapshot functions through the owner's mapped write repository. Preserve the database revision alongside the restored object.
4. Inject that repository into the named command handler and load its target. Distinguish a missing root from corrupt stored state.
5. Invoke aggregate behaviour and save explicitly. A read/restoration operation alone must neither change a revision nor publish domain/realtime events.
6. Test round-trip restoration, malformed state, identity mismatch, empty pending facts and a subsequent valid transition. Use the [transaction recipe](transactions.md) to prove rollback and repository scope.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## The mappings in three languages

The mapped write repository preserves existence and revision while replacing the
stored representation with a restored domain object. The owner-specific binding
on the [write repository page](write-repositories.md) selects the restoration function.

### Go

Source: [mapped_write_repository.go](../../../services/storefront/adaptors/persistence/postgres/mapped_write_repository.go).

```go source="../../../services/storefront/adaptors/persistence/postgres/mapped_write_repository.go"
func (r MappedWriteRepository[S, A]) Get(ctx context.Context, id string) (a.Loaded[A], error) {
	loaded, err := r.Source.Get(ctx, id)
	result := a.Loaded[A]{Exists: loaded.Exists, Version: loaded.Version}
	if err != nil || !loaded.Exists {
		return result, err
	}
	result.State, err = r.Restore(loaded.State)
	return result, err
}
```

### Python

Source: [mapped_write_repository.py](../../../services/operations/src/operations/adaptors/persistence/mapped_write_repository.py).

```python source="../../../services/operations/src/operations/adaptors/persistence/mapped_write_repository.py"
def get(self, identity: str) -> Loaded[A] | None:
    loaded = self._source.get(identity)
    return (
        Loaded(exists=True, version=loaded["version"], state=self._restore(loaded["state"]))
        if loaded
        else None
    )
```

### TypeScript

Source: [mapped-write-repository.ts](../../../services/engagement/src/adaptors/persistence/mapped-write-repository.ts).

```ts source="../../../services/engagement/src/adaptors/persistence/mapped-write-repository.ts"
async get(id: string) {
  const loaded = await this.source.get(id);
  return loaded ? { ...loaded, state: this.restore(loaded.state) } : undefined;
}
```

The handler's `loaded.state` is therefore already hydrated. It should not decode
JSON, call a transport codec or restore state a second time.

## Restoration is a validation boundary

Order's `Restore` reconstructs its lines and checks identity, status and total
quantity. PreparationTicket's `restore` converts persisted field names, validates
the lifecycle and classifies invalid stored state as `CorruptState`. Reward's
constructor validates terms and redemption state before copying its saved value.

### Python

Source: [preparation_ticket.py](../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py).

```python source="../../../services/operations/src/operations/contexts/preparation/domain/preparation_ticket.py"
@classmethod
def restore(cls, value: object) -> "PreparationTicket":
    try:
        fields = record(value)
        match fields["status"]:
            case "queued" | "preparing" | "ready" as status:
                return cls(
                    TicketState(
                        identifier(fields["id"]),
                        identifier(fields["orderId"]),
                        identifier(fields["customerId"]),
                        text(fields["instructions"]),
                        status,
                    )
                )
            case _:
                raise ValueError("Invalid preparation status")
    except (Rejection, KeyError, TypeError, ValueError) as error:
        raise CorruptState("Corrupt preparation ticket state") from error
```

See the Go and TypeScript restoration bodies in
[Order](../../../services/storefront/contexts/ordering/domain/order.go) and [Reward](../../../services/engagement/src/contexts/loyalty/domain/reward.ts).
Restoration starts with an empty pending fact collection. It never invokes the
creation or mutation operations that would raise a fresh business occurrence.

## Hydration and event sourcing

An event-sourced aggregate would fold an authoritative ordered event stream,
possibly starting at a cached snapshot. That is a different persistence choice.
The [event-sourcing guide](../event-sourcing.md) explains why replaying an outbox
delivery does not perform aggregate rehydration here.

## Related concepts

[Write repositories](write-repositories.md) · [Factories](factories.md) · [Aggregates](aggregates.md) · [Read repositories](read-repositories.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
