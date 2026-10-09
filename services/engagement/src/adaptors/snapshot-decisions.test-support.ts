import type {Change, Loaded, Metadata} from '../foundation/application.js';
import {PostgresContextDatabase} from './postgres.js';
import {PostgresAggregateTransaction} from './aggregate-transaction.js';

/** Technical fixtures drive the real command transaction with snapshot-only decisions. */
export class SnapshotDecisionFixture<S> {
  private transaction: PostgresAggregateTransaction<S, S>;
  constructor(database: PostgresContextDatabase, kind: string, restore: (value: unknown) => S) {
    this.transaction = new PostgresAggregateTransaction(database, kind, restore, repository => repository);
  }
  execute(metadata: Metadata, decide: (loaded: Loaded<S> | undefined) => Change<S>) {
    return this.transaction.execute(metadata, async repository => {
      const change = decide(await repository.get(metadata.target));
      if (change.changed) await repository.save(change.state);
      return {status: change.status, publications: change.publications ?? []};
    });
  }
}
