# Transactions and optimistic concurrency

The central command executor owns the transaction for one aggregate invocation.
It records the business outcome and commits state together with receipts,
outgoing private-event intent, realtime intent and API reply bytes where applicable.

## Creation recipe

1. Configure one owner database, aggregate kind, restore function, mapped write repository and private-fact mapper in composition.
2. Register plain command factories through the shared registration helper. It supplies the executor and each invocation's repository and realtime bus.
3. Keep load/mutate/save in the named handler. Explicitly publish browser messages when needed; save itself produces no realtime intent.
4. Let the executor perform receipt/identity checks, expected-version validation, final flush and outgoing encoding. Commit one root, outcome, receipts, selected private/realtime bytes and exact API reply together.
5. An expected rejection records its outcome without flushing staged state. A matching explicitly selected rejection message may commit with it. Technical failures roll back and remain retryable.
6. For event-based integration/realtime output, use its dedicated subscription bus. It commits the receiving receipt and outgoing bytes separately from the original root transition, before ACK.
7. Test absent intent, explicit intent, encoding/receipt failures, identical retries, conflicting identities, expired repository/bus scope and the real broker/browser recovery paths.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Execution boundary in three languages

The following excerpts show the expected-version check and the event-recording
repository around the feature invocation. The surrounding implementation owns
connection lifetime, locking, commit and rollback.

### Go

Source: [aggregate_transaction.go](../../../services/storefront/adaptors/persistence/postgres/aggregate_transaction.go).

```go source="../../../services/storefront/adaptors/persistence/postgres/aggregate_transaction.go"
if m.ExpectedVersion != nil && *m.ExpectedVersion != loaded.Version {
	detail := (&a.ApplicationError{Code: "version_conflict", Message: "The expected aggregate version is stale"}).Rejection()
	outcome.Rejection = &detail
} else {
	ownerRepository := &execution.EventRecordingRepository[A]{Source: s.repository(repository), Map: s.publications}
	result, err = work(ownerRepository)
	result.Publications = append(result.Publications, ownerRepository.Publications...)
	repository.close()
	if err != nil {
		var expected d.ExpectedError
		if !errors.As(err, &expected) {
			return a.Outcome{}, err
		}
		detail := expected.Rejection()
		outcome.Rejection = &detail
	} else {
		outcome.Status = result.Status
		if err = s.persistResult(ctx, tx, m, repository, result, &outcome); err != nil {
			return a.Outcome{}, err
		}
	}
}
```

### Python

Source: [aggregate_transaction.py](../../../services/operations/src/operations/adaptors/persistence/aggregate_transaction.py).

```python source="../../../services/operations/src/operations/adaptors/persistence/aggregate_transaction.py"
                if m.expected is not None and m.expected != version:
                    raise VersionConflictApplicationError()
                try:
                    owner_repository = EventRecordingWriteRepository(
                        self.repository(repository), self.publications
                    )
                    result = work(owner_repository)
                    result = replace(result, publications=result.publications + owner_repository.publications)
                finally:
                    repository.close()
                outcome["status"] = result.status
                if result.rejection is not None:
                    outcome["rejection"] = result.rejection
            except Rejection as rejection:
                repository.close()
                outcome["rejection"] = rejection.outcome()
            else:
                if result.rejection is None:
                    outcome["version"] = self._persist_result(connection, m, repository, result)
```

### TypeScript

Source: [aggregate-transaction.ts](../../../services/engagement/src/adaptors/persistence/aggregate-transaction.ts).

```ts source="../../../services/engagement/src/adaptors/persistence/aggregate-transaction.ts"
        if (m.expected !== undefined && m.expected !== version) throw new VersionConflictApplicationError();
        const ownerRepository = new EventRecordingWriteRepository(
          this.repository(repository),
          this.publications,
        );
        result = await work(ownerRepository);
        result = {
          ...result,
          publications: [...(result.publications ?? []), ...ownerRepository.publications],
        };
        outcome.status = result.status;
        if (result.rejection) outcome.rejection = result.rejection;
      } catch (error) {
        if (!(error instanceof Rejection)) throw error;
        outcome.rejection = error.outcome();
      } finally {
        repository.close();
      }
      if (result && !result.rejection) await this.persistResult(client, m, repository, result, outcome);
      await persistRealtime(
        client,
        this.db.owner,
        this.kind,
        m,
        outcome,
        repository.pending,
        result?.realtime ?? [],
      );
```

The command handler has no transaction callback. Its repository stages state.
The executor owns whether all that work becomes durable together.

## Concurrency and retry identity

The context transaction serialises the target, including creation before its row
exists. An expected revision ensures a caller-driven decision is based on the
state the caller observed. A stale fresh command receives `version_conflict`.

An identical retry is recognised from its existing receipt before a new decision
is attempted. It can recover its old outcome after another command has advanced
the aggregate. Reusing the command ID with changed material input is an idempotency
conflict. Correlation IDs only group workflow evidence.

## Failure boundary

Expected business refusals are recorded outcomes. Technical or corrupt-state
failures roll back and remain retryable. A crash after commit but before reply or
ACK is recovered through receipts and stored outgoing bytes. A crash before
commit leaves the invocation available for retry.

Only real database and broker tests establish these claims. The
[testing guide](../testing-and-exercises.md) explains which evidence proves each boundary.

## Related concepts

[Command handlers](command-handlers.md) · [Write repositories](write-repositories.md) · [Outbox](outbox.md) · [Receipts and idempotency](receipts-and-idempotency.md) · [Realtime events](realtime-events.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
