import type {PoolClient} from 'pg';
import {createHash} from 'node:crypto';
import {VersionConflictApplicationError, type Metadata, type Outcome, type Publication} from '../foundation/application.js';
import type {WriteRepository} from '../foundation/write-repository.js';
import {type AggregateTransaction, type TransactionResult, EventRecordingWriteRepository} from './command-execution.js';
import {identifier, Rejection} from '../foundation/domain.js';
import {PostgresContextDatabase} from './postgres.js';
import {PostgresSnapshotWriteRepository} from './snapshot-write-repository.js';
import {encode, realtime} from './codec.js';
import {commitOutcome} from './replies.js';
import {outcome as decodeOutcome} from './receipts.js';

function canonical(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(canonical);
  if (value && typeof value === 'object') return Object.fromEntries(
    Object.entries(value).sort(([a], [b]) => a.localeCompare(b)).map(([key, item]) => [key, canonical(item)]));
  return value;
}
/** Commits one aggregate, command outcome, receipts and outgoing event/realtime intent atomically. */
export class PostgresAggregateTransaction<S, A> implements AggregateTransaction<A> {
  constructor(private db: PostgresContextDatabase, private kind: string, private restore: (value: unknown) => S,
    private repository: (source: WriteRepository<S>) => WriteRepository<A>, private publications: (aggregate: A) => Publication[] = () => []) {}
  async execute(m: Metadata, work: (repository: WriteRepository<A>) => Promise<TransactionResult>): Promise<Outcome> {
    [m.id, m.target, m.correlation].forEach(identifier);
    const digest = createHash('sha256').update(JSON.stringify(canonical({expected: m.expected ?? null, input: m.input}))).digest('hex');
    const target = this.kind+':'+m.target, client = await this.db.pool.connect();
    try {
      await client.query('BEGIN');
      await client.query('SELECT pg_advisory_xact_lock(hashtextextended($1,0))', [target]);
      await client.query("SELECT set_config('cafe.command_target',$1,true)", [target]);
      if (m.consumer) {
        await client.query('SELECT pg_advisory_xact_lock(hashtextextended($1,0))', ['consumer:'+m.consumer+':'+m.sourceId]);
        const {rows: [previous]} = await client.query(`SELECT fingerprint,target,outcome FROM cafe.consumer_receipts
          WHERE consumer=$1 AND event_id=$2`, [m.consumer, m.sourceId]);
        if (previous) {
          if (previous.fingerprint !== m.sourceHash || previous.target !== target) throw new Error('Conflicting delivery identity');
          const saved = decodeOutcome(previous.outcome, m.target);
          return await commitOutcome(client, saved);
        }
      }
      const {rows: [receipt]} = await client.query(`SELECT fingerprint,outcome FROM cafe.command_receipts
        WHERE kind=$1 AND aggregate_id=$2 AND command_name=$3 AND command_id=$4`, [this.kind, m.target, m.name, m.id]);
      if (receipt) {
        if (receipt.fingerprint !== digest) {
          return await commitOutcome(client, {aggregateId: m.target, version: 0, status: '', rejection: {
            code: 'idempotency_conflict', message: 'The command identity has different input'}});
        }
        const saved = decodeOutcome(receipt.outcome, m.target);
        await incoming(client, m, target, saved);
        return await commitOutcome(client, saved);
      }
      const repository = new PostgresSnapshotWriteRepository(client, this.kind, m.target, this.restore);
      await repository.load();
      const version = repository.version;
      const outcome: Outcome = {aggregateId: m.target, version, status: ''};
      let result: TransactionResult | undefined;
      try {
        if (m.expected !== undefined && m.expected !== version) throw new VersionConflictApplicationError();
        const ownerRepository = new EventRecordingWriteRepository(this.repository(repository), this.publications);
        result = await work(ownerRepository);
        result = {...result, publications: [...result.publications ?? [], ...ownerRepository.publications]};
        outcome.status = result.status;
      } catch (error) {
        if (!(error instanceof Rejection)) throw error;
        outcome.rejection = error.outcome();
      }
      finally { repository.close(); }
      if (result) await this.persistResult(client, m, repository, result, outcome);
      await client.query(`INSERT INTO cafe.command_receipts(kind,aggregate_id,command_name,command_id,
        fingerprint,outcome,correlation_id,causation_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
        [this.kind, m.target, m.name, m.id, digest, outcome, m.correlation, m.causation ?? null]);
      await incoming(client, m, target, outcome);
      return await commitOutcome(client, outcome);
    } catch (error) {
      await client.query('ROLLBACK').catch(() => {});
      throw error;
    } finally { client.release(); }
  }

  private async persistResult(client: PoolClient, m: Metadata, repository: PostgresSnapshotWriteRepository<S>, result: TransactionResult, outcome: Outcome) {
    if (repository.pending === undefined) {
      if (result.publications?.length) throw new Error('A no-op cannot publish new events');
      return;
    }
    outcome.version = await repository.flush();
    for (const publication of result.publications ?? []) {
      const {event, body} = encode(this.db.owner, this.kind, m.target, outcome.version, m, publication);
      await client.query(`INSERT INTO cafe.outbox_events(id,event_name,visibility,aggregate_kind,aggregate_id,
        aggregate_version,correlation_id,causation_id,body) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
        [event.id, event.name, event.visibility, this.kind, m.target, outcome.version, m.correlation, m.id, body]);
      await client.query('INSERT INTO cafe.dispatches(event_id) VALUES($1)', [event.id]);
    }
    const publication = realtime(this.db.owner, this.kind, m.target, outcome.version, repository.pending);
    await client.query(`INSERT INTO cafe.realtime_publications(id,channel,aggregate_kind,aggregate_id,revision,body)
      VALUES($1,$2,$3,$4,$5,$6)`, [publication.id, 'cafe:'+this.db.owner, this.kind, m.target, outcome.version, publication.body]);
    await client.query('INSERT INTO cafe.realtime_dispatches(event_id) VALUES($1)', [publication.id]);
  }
}
async function incoming(client: PoolClient, m: Metadata, target: string, outcome: Outcome) {
  if (m.consumer) await client.query(`INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome)
    VALUES($1,$2,$3,$4,$5)`, [m.consumer, m.sourceId, m.sourceHash, target, outcome]);
}
