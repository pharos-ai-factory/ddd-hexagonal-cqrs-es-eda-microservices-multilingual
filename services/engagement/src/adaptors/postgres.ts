import {Pool, type PoolClient} from 'pg';
import {createHash} from 'node:crypto';
import {VersionConflictApplicationError, type Change, type AggregateCommandPort, type Loaded, type Metadata, type Outcome, type QueryPort} from '../foundation/application.js';
import type {Page, PageRequest} from '../foundation/pagination.js';
import {identifier, Rejection} from '../foundation/domain.js';
import {encode, realtime} from './codec.js';
import {commitOutcome} from './replies.js';
import {outcome as decodeOutcome, checkRootIdentity} from './receipts.js';
import schema from './generated/persistence.json' with {type: 'json'};
import contextSchemas from './generated/context-persistence.json' with {type: 'json'};

function canonical(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(canonical);
  if (value && typeof value === 'object') return Object.fromEntries(
    Object.entries(value).sort(([a], [b]) => a.localeCompare(b)).map(([key, item]) => [key, canonical(item)]));
  return value;
}
/** Owns one context pool and verifies its database identity, grants and migration ledger. */
export class PostgresContextDatabase {
  readonly pool: Pool;
  constructor(readonly owner: string, url: string) {
    this.pool = new Pool({connectionString: url, max: 6, connectionTimeoutMillis: 3000});
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
    const {rows: [identity]} = await this.pool.query('SELECT owner FROM cafe.context_identity WHERE singleton');
    const expected = (contextSchemas as Record<string, Record<string, string>>)[this.owner];
    if (!expected || identity?.owner !== this.owner) throw new Error('Context schema identity mismatch');
    const {rows: migrations} = await this.pool.query<{version: number; checksum: string}>(
      'SELECT version,checksum FROM cafe.context_migrations');
    if (migrations.length !== Object.keys(expected).length || migrations.some(item => expected[String(item.version)] !== item.checksum)) {
      throw new Error('Context migration checksum mismatch');
    }
    await this.pool.query('SELECT id FROM cafe.realtime_publications LIMIT 0');
  }
}
/** Commits one aggregate, command outcome, receipts and outgoing event/realtime intent atomically. */
export class PostgresAggregateCommandStore<S> implements AggregateCommandPort<S> {
  constructor(private db: PostgresContextDatabase, private kind: string, private restore: (value: unknown) => S) {}
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
      const {rows: [row]} = await client.query('SELECT version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2 FOR UPDATE',
        [this.kind, m.target]);
      const version = Number(row?.version ?? 0);
      const loaded: Loaded<S> | undefined = row ? {exists: true, version, state: this.restore(row.state)} : undefined;
      if (loaded) checkRootIdentity(loaded.state, m.target);
      const outcome: Outcome = {aggregateId: m.target, version, status: ''};
      let change: Change<S> | undefined;
      try {
        if (m.expected !== undefined && m.expected !== version) throw new VersionConflictApplicationError();
        change = decide(loaded);
        outcome.status = change.status;
      } catch (error) {
        if (!(error instanceof Rejection)) throw error;
        outcome.rejection = error.outcome();
      }
      if (change) await this.persistChange(client, m, Boolean(row), change, outcome);
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

  private async persistChange(client: PoolClient, m: Metadata, exists: boolean, change: Change<S>, outcome: Outcome) {
    if (!change.changed) {
      if (change.publications?.length) throw new Error('A no-op cannot publish new events');
      return;
    }
    checkRootIdentity(change.state, m.target);
    outcome.version++;
    if (exists) await client.query('UPDATE cafe.aggregates SET version=$3,state=$4 WHERE kind=$1 AND id=$2',
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
  }
}
async function incoming(client: PoolClient, m: Metadata, target: string, outcome: Outcome) {
  if (m.consumer) await client.query(`INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome)
    VALUES($1,$2,$3,$4,$5)`, [m.consumer, m.sourceId, m.sourceHash, target, outcome]);
}
/** Restores validated snapshots through the read port without exposing database objects. */
export class PostgresAggregateQueries<S> implements QueryPort<S> {
  constructor(private db: PostgresContextDatabase, private kind: string, private restore: (value: unknown) => S) {}
  private loaded(row: {id: string; version: number; state: unknown}): Loaded<S> {
    const state = this.restore(row.state);
    checkRootIdentity(state, row.id);
    return {exists: true, version: Number(row.version), state};
  }
  async get(id: string): Promise<Loaded<S> | undefined> {
    identifier(id);
    const {rows: [row]} = await this.db.pool.query('SELECT id,version,state FROM cafe.aggregates WHERE kind=$1 AND id=$2', [this.kind, id]);
    return row ? this.loaded(row) : undefined;
  }
  async page(request: PageRequest): Promise<Page<S>> {
    if (!Number.isInteger(request.limit) || request.limit < 1 || request.limit > 100) throw new Error('Invalid page size');
    if (request.after) identifier(request.after);
    const {rows} = await this.db.pool.query(
      'SELECT id,version,state FROM cafe.aggregates WHERE kind=$1 AND ($2::uuid IS NULL OR id>$2::uuid) ORDER BY id LIMIT $3',
      [this.kind, request.after ?? null, request.limit+1]);
    const selected = rows.slice(0, request.limit);
    return {items: selected.map(row => this.loaded(row)),
      nextId: rows.length > request.limit ? selected.at(-1)!.id as string : undefined};
  }
  async list(): Promise<Loaded<S>[]> {
    const {rows} = await this.db.pool.query('SELECT id,version,state FROM cafe.aggregates WHERE kind=$1 ORDER BY id', [this.kind]);
    return rows.map(row => this.loaded(row));
  }
}
