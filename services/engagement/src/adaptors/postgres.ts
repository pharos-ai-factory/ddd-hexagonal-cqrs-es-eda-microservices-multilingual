import {Pool, type PoolClient} from 'pg';
import {createHash} from 'node:crypto';
import type {Change, CommandPort, Loaded, Metadata, Outcome, QueryPort} from '../foundation/application.js';
import {identifier, Rejection} from '../foundation/domain.js';
import {encode, realtime} from './codec.js';
import {outcome as decodeOutcome, checkRootIdentity} from './receipts.js';
import schema from './generated/persistence.json' with {type: 'json'};

function canonical(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(canonical);
  if (value && typeof value === 'object') return Object.fromEntries(
    Object.entries(value).sort(([a], [b]) => a.localeCompare(b)).map(([key, item]) => [key, canonical(item)]));
  return value;
}
export class Database {
  readonly pool: Pool;
  constructor(readonly owner: string, url: string) {
    this.pool = new Pool({connectionString: url, max: 6});
    // The pool removes failed idle clients; queries can acquire a replacement.
    this.pool.on('error', () => console.warn('Idle PostgreSQL connection lost for', this.owner));
  }
  async verify() {
    const {rows: [row]} = await this.pool.query(`SELECT current_database() AS database,
      rolsuper OR rolcreatedb OR rolcreaterole OR has_database_privilege(current_user,current_database(),'CREATE')
      OR has_schema_privilege(current_user,'cafe','CREATE') AS privileged
      FROM pg_roles WHERE rolname=current_user`);
    if (row.database !== 'cafe_'+this.owner || row.privileged) throw new Error('Runtime owner or privilege mismatch');
    for (const [version, table] of [[1, 'schema_version'], [2, 'schema_migrations']] as const) {
      const {rows: [migration]} = await this.pool.query(`SELECT checksum FROM cafe.${table} WHERE version=$1`, [version]);
      if (migration?.checksum !== schema[String(version) as keyof typeof schema]) throw new Error('Database migration checksum mismatch');
    }
    await this.pool.query('SELECT id FROM cafe.realtime_publications LIMIT 0');
  }
}
export class Commands<S> implements CommandPort<S> {
  constructor(private db: Database, private kind: string) {}
  async execute(m: Metadata, decide: (loaded: Loaded<S> | undefined) => Change<S>): Promise<Outcome> {
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
          await client.query('COMMIT'); return saved;
        }
      }
      const {rows: [receipt]} = await client.query(`SELECT fingerprint,outcome FROM cafe.command_receipts
        WHERE kind=$1 AND aggregate_id=$2 AND command_name=$3 AND command_id=$4`, [this.kind, m.target, m.name, m.id]);
      if (receipt) {
        if (receipt.fingerprint !== digest) {
          await client.query('ROLLBACK');
          return {aggregateId: m.target, version: 0, status: '', rejection: {
            code: 'idempotency_conflict', message: 'The command identity has different input'}};
        }
        const saved = decodeOutcome(receipt.outcome, m.target);
        await incoming(client, m, target, saved);
        await client.query('COMMIT'); return saved;
      }
      const {rows: [row]} = await client.query('SELECT version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2 FOR UPDATE',
        [this.kind, m.target]);
      const version = Number(row?.version ?? 0);
      if (row) checkRootIdentity(row.state, m.target);
      const outcome: Outcome = {aggregateId: m.target, version, status: ''};
      let change: Change<S> | undefined;
      try {
        if (m.expected !== undefined && m.expected !== version) throw new Rejection('version_conflict', 'The expected version is stale');
        change = decide(row ? {exists: true, version, state: row.state as S} : undefined);
        outcome.status = change.status;
      } catch (error) {
        if (!(error instanceof Rejection)) throw error;
        outcome.rejection = error.outcome();
      }
      if (change?.changed) {
        checkRootIdentity(change.state, m.target);
        outcome.version++;
        if (row) await client.query('UPDATE cafe.aggregates SET version=$3,state=$4 WHERE kind=$1 AND id=$2',
          [this.kind, m.target, outcome.version, JSON.stringify(change.state)]);
        else await client.query('INSERT INTO cafe.aggregates(kind,id,version,state) VALUES($1,$2,$3,$4)',
          [this.kind, m.target, outcome.version, JSON.stringify(change.state)]);
        for (const publication of change.publications ?? []) {
          const {event, body} = encode(this.db.owner, this.kind, m.target, outcome.version, m, publication);
          await client.query(`INSERT INTO cafe.outbox_events(id,event_name,visibility,aggregate_kind,aggregate_id,
            aggregate_version,correlation_id,causation_id,body) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
            [event.id, event.name, event.visibility, this.kind, m.target, outcome.version, m.correlation, m.id, body]);
          await client.query('INSERT INTO cafe.dispatches(event_id) VALUES($1)', [event.id]);
        }
        const publication = realtime(this.db.owner, this.kind, m.target, outcome.version, change.state);
        await client.query(`INSERT INTO cafe.realtime_publications(id,channel,aggregate_kind,aggregate_id,revision,body)
          VALUES($1,$2,$3,$4,$5,$6)`, [publication.id, 'cafe:'+this.db.owner, this.kind, m.target, outcome.version, publication.body]);
        await client.query('INSERT INTO cafe.realtime_dispatches(event_id) VALUES($1)', [publication.id]);
      } else if (change?.publications?.length) throw new Error('A no-op cannot publish new events');
      await client.query(`INSERT INTO cafe.command_receipts(kind,aggregate_id,command_name,command_id,
        fingerprint,outcome,correlation_id,causation_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
        [this.kind, m.target, m.name, m.id, digest, outcome, m.correlation, m.causation ?? null]);
      await incoming(client, m, target, outcome);
      await client.query('COMMIT');
      return outcome;
    } catch (error) {
      await client.query('ROLLBACK').catch(() => {});
      throw error;
    } finally { client.release(); }
  }
}
async function incoming(client: PoolClient, m: Metadata, target: string, outcome: Outcome) {
  if (m.consumer) await client.query(`INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome)
    VALUES($1,$2,$3,$4,$5)`, [m.consumer, m.sourceId, m.sourceHash, target, outcome]);
}
export class Queries<S> implements QueryPort<S> {
  constructor(private db: Database, private kind: string) {}
  async get(id: string): Promise<Loaded<S> | undefined> {
    identifier(id);
    const {rows: [row]} = await this.db.pool.query('SELECT version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2', [this.kind, id]);
    if (row) checkRootIdentity(row.state, id);
    return row ? {exists: true, version: Number(row.version), state: row.state as S} : undefined;
  }
  async list(): Promise<Loaded<S>[]> {
    const {rows} = await this.db.pool.query('SELECT id,version,state FROM cafe.aggregates WHERE kind=$1 ORDER BY id LIMIT 100', [this.kind]);
    return rows.map(row => {
      checkRootIdentity(row.state, row.id);
      return {exists: true, version: Number(row.version), state: row.state as S};
    });
  }
}
