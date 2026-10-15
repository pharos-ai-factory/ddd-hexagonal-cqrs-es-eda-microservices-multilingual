# Receipts and idempotency

Idempotency preserves an operation's intended effect when it is repeated.
Durable receipts record the identity, material input and outcome needed to
recognise work already accepted or completed.

## Keep identities separate

| Identity                  | What it identifies                 | Use                                             |
| ------------------------- | ---------------------------------- | ----------------------------------------------- |
| Command ID                | One business attempt               | Recover its stored outcome                      |
| Event ID                  | One published occurrence           | Detect repeated delivery to a subscriber        |
| Business key              | The underlying business occurrence | Prevent a second credit for one collected order |
| Aggregate ID and revision | One root and its committed state   | Target and concurrency checks                   |
| Correlation ID            | A workflow group                   | Trace related work                              |
| Causation ID              | The triggering message             | Explain why the next work exists                |

One correlation can cover many valid changes, so it cannot serve as a deduplication key.

## Conflicting command reuse in three languages

These excerpts compare the stored input fingerprint before recovering an outcome:

### Go

Source: [aggregate_transaction.go](../../../services/storefront/adaptors/persistence/postgres/aggregate_transaction.go).

```go source="../../../services/storefront/adaptors/persistence/postgres/aggregate_transaction.go"
if savedHash != hash {
	detail := (&a.ApplicationError{Code: "idempotency_conflict", Message: "The command identity was reused with different input"}).Rejection()
	return commitOutcome(ctx, tx, a.Outcome{AggregateID: m.AggregateID, Rejection: &detail})
}
```

### Python

Source: [aggregate_transaction.py](../../../services/operations/src/operations/adaptors/persistence/aggregate_transaction.py).

```python source="../../../services/operations/src/operations/adaptors/persistence/aggregate_transaction.py"
if receipt["fingerprint"] != digest:
    return append_reply(
        connection,
        {
            "aggregateId": m.target,
            "version": 0,
            "status": "",
            "rejection": {
                "code": "idempotency_conflict",
                "message": "The command identity has different input",
            },
        },
    )
```

### TypeScript

Source: [aggregate-transaction.ts](../../../services/engagement/src/adaptors/persistence/aggregate-transaction.ts).

```ts source="../../../services/engagement/src/adaptors/persistence/aggregate-transaction.ts"
if (receipt.fingerprint !== digest) {
  return await commitOutcome(client, {
    aggregateId: m.target,
    version: 0,
    status: '',
    rejection: {
      code: 'idempotency_conflict',
      message: 'The command identity has different input',
    },
  });
}
```

The fingerprint includes the original expected version and material input. An
identical retry recovers its previous outcome even if the root later advances.
A new intended decision uses a fresh command identity.

## Receipt atomicity

Recording a receipt before the owned work commits can suppress unfinished work.
Recording it after acknowledging leaves a duplicate window. Commit the receipt
with the owned state or queued hand-off, then ACK that delivery.

For publication subscribers, that transaction contains the incoming receipt and
outgoing integration bytes. For aggregate-changing event reactions, receiving
evidence accompanies the exact private command and dispatch intent. The later
command has its own execution receipt and outcome.

## External effects

A local receipt cannot atomically include an unrelated provider's database.
Communication uses an idempotent provider request and retained acceptance evidence
before recording delivery. The [consistency guide](../consistency-and-delivery.md)
explains the limits of exactly-once claims across independent systems.

## Related concepts

[Commands](commands.md) · [Transactions](transactions.md) · [Durable commands](durable-commands.md) · [Outbox](outbox.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
