# Transactional outbox

A transactional outbox records outgoing intent in the same database transaction
as the local work that requires it. A relay delivers the committed bytes later.
This closes the failure window between committing state and asking a broker to publish.

## Outgoing paths in this reference

| Local work                        | Stored outgoing intent                              | Later delivery                |
| --------------------------------- | --------------------------------------------------- | ----------------------------- |
| Aggregate transition              | Private facts and explicitly selected browser bytes | RabbitMQ and realtime gateway |
| Private publication subscriber    | Integration bytes plus subscriber receipt           | RabbitMQ                      |
| Aggregate-changing event reaction | Private command bytes plus receiving evidence       | Owner command queue           |
| API command execution             | Exact reply bytes                                   | Request/reply transport       |

Each row has its own transaction. The integration publication transaction follows
private domain delivery; it is separate from the original aggregate commit.

## Three implementations of the publication hand-off

These excerpts show the encoding or message construction using source metadata.
The linked complete implementations insert the bytes, dispatch and receipt and
commit before returning to the consumer.

### Go

Source: [publication_outbox.go](../../../services/storefront/adaptors/persistence/postgres/publication_outbox.go).

```go source="../../../services/storefront/adaptors/persistence/postgres/publication_outbox.go"
message := a.Message{ID: NewID(), Name: publication.Name, Context: p.db.owner, Visibility: a.Public, ContractVersion: 1, AggregateKind: p.kind, AggregateID: m.AggregateID, AggregateVersion: m.SourceVersion, CorrelationID: m.CorrelationID, CausationID: m.SourceEventID, OccurredAt: m.SourceTime, Payload: publication.Payload}
if err = p.db.appendEvent(ctx, tx, message); err != nil {
	return a.Outcome{}, err
}
outcome = a.Outcome{AggregateID: m.AggregateID, Version: m.SourceVersion, Status: "publication_queued"}
if err = recordIncoming(ctx, tx, m, target, outcome); err != nil {
	return a.Outcome{}, err
}
```

### Python

Source: [integration_event_outbox.py](../../../services/operations/src/operations/adaptors/persistence/integration_event_outbox.py).

```python source="../../../services/operations/src/operations/adaptors/persistence/integration_event_outbox.py"
publication = self.publication(payload)
if publication.name not in PAYLOAD_CODECS:
    raise ValueError("Publication handler must stage an integration event")
event, body = codec.encode(
    self.database.owner,
    self.kind,
    m.target,
    m.source_version,
    replace(m, id=m.source_id),
    publication,
    occurred_at=m.source_time,
)
```

### TypeScript

Source: [integration-event-outbox.ts](../../../services/engagement/src/adaptors/persistence/integration-event-outbox.ts).

```ts source="../../../services/engagement/src/adaptors/persistence/integration-event-outbox.ts"
const { event, body } = encode(
  this.db.owner,
  this.kind,
  m.target,
  m.sourceVersion,
  { ...m, id: m.sourceId! },
  this.publication(payload),
  m.sourceTime,
);
if (event.visibility !== 'integration') throw new Error('Publication handler requires an integration event');
```

## Relay and recovery

The relay claims pending work using a lease and fencing token, publishes using
broker confirms and records completion. A crash after publication can cause
redelivery. A stale relay cannot finish a newer owner's claim. Consumer receipts
and business keys handle duplicates at the appropriate boundary.

Stored exact bytes prevent a retry from silently re-encoding an old intent under
new application state. Schema evolution still needs compatibility checks because
the receiver may be newer than the code that originally encoded those bytes.

An outbox is a delivery record. It does not promise a complete authoritative
event history for [event sourcing](../event-sourcing.md).

## Event-driven browser publication

A declared realtime subscriber commits its receipt and exact browser bytes in its
own transaction. Saving an aggregate alone creates no browser intent. Follow the
[realtime recipe](realtime-events.md) for the three language implementations.

## Related concepts

[Transactions](transactions.md) · [Domain events](domain-events.md) · [Integration events](integration-events.md) · [Receipts and idempotency](receipts-and-idempotency.md) · [Realtime events](realtime-events.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
