import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {readFileSync} from 'node:fs';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
import pg from 'pg';
import amqp from 'amqplib';
import protobuf from 'protobufjs';
import {eventually, type Loaded} from './client.js';

const sourceMap = JSON.parse(execFileSync('python3', ['scripts/contract_sources.py'], {encoding: 'utf8'})) as
  Record<string, {entrypoint: string; sources: Record<string, string>}>;
const group = sourceMap.events!;
const eventSchema = new protobuf.Root();
const files = Object.fromEntries(Object.entries(group.sources).map(([logical, physical]) => [logical, path.resolve(physical)]));
const physical = new Set(Object.values(files));
eventSchema.resolvePath = (_origin, target) => physical.has(target) ? target : files[target]!;
eventSchema.loadSync(files[group.entrypoint]!);
eventSchema.resolveAll();
const eventType = eventSchema.lookupType('cafe.v1.Event');
const catalogue = ['menu','ordering','preparation','collection','loyalty'].flatMap(owner =>
  JSON.parse(readFileSync(`contracts/${owner}/messaging/integration_events/v1/catalogue.json`, 'utf8')) as
  {name: string; owner: string; consumer: string; visibility: string}[]);
// These test controls describe private queue routing; private payloads remain opaque.
const privateDeliveries = [
  {name: 'loyalty.reward-earned', owner: 'loyalty', consumer: 'loyalty.issue-reward', visibility: 'domain'},
  {name: 'communication.notification-requested', owner: 'communication', consumer: 'communication.deliver-notice', visibility: 'domain'},
];
// Field 1 is the immutable publication ID in the existing delivery mechanism.
// Replace only that length-delimited header and preserve every private payload byte.
function freshIdentity(bytes: Buffer, id: string): Buffer {
  assert.equal(bytes[0], 10);
  assert.equal(bytes[1], 36);
  return Buffer.concat([bytes.subarray(0, 2), Buffer.from(id), bytes.subarray(38)]);
}

export class Evidence {
  private pools = new Map<string, pg.Pool>();
  constructor(private values: Record<string, string>) {}
  pool(owner: string): pg.Pool {
    assert.ok(['menu', 'ordering', 'preparation', 'collection', 'loyalty', 'communication'].includes(owner));
    let pool = this.pools.get(owner);
    if (!pool) {
      pool = new pg.Pool({host: '127.0.0.1', port: Number(this.values.PG_PORT), database: 'cafe_'+owner,
        user: 'cafe_'+owner, password: this.values[owner.toUpperCase()+'_DB_PASSWORD'], max: 4,
        connectionTimeoutMillis: 5_000, query_timeout: 10_000});
      this.pools.set(owner, pool);
    }
    return pool;
  }
  async close() { await Promise.all([...this.pools.values()].map(pool => pool.end())); }
  async roots(owner: string, kind: string, field: string, identity: string): Promise<Loaded[]> {
    const result = await this.pool(owner).query(
      'SELECT version,state FROM cafe.aggregates WHERE kind=$1 AND state->>$2=$3 ORDER BY id', [kind, field, identity]);
    return result.rows.map(row => ({version: Number(row.version), state: row.state}));
  }
  async events(owner: string, name: string, aggregate: string) {
    return (await this.pool(owner).query(
      'SELECT id,body FROM cafe.outbox_events WHERE event_name=$1 AND aggregate_id=$2 ORDER BY created_at,id',
      [name, aggregate])).rows as {id: string; body: Buffer}[];
  }
  async receipt(consumer: string, id: string) {
    const owner = consumer.split('.')[0]!;
    return eventually('committed receipt for '+consumer, async () => {
      const result = await this.pool(owner).query('SELECT outcome FROM cafe.consumer_receipts WHERE consumer=$1 AND event_id=$2', [consumer, id]);
      if (!result.rows[0]) return undefined;
      assert.equal(result.rows[0].outcome.rejection, undefined);
      return result.rows[0].outcome;
    });
  }
  async redeliver(name: string, aggregate: string): Promise<{id: string; consumer: string}> {
    const entry = [...catalogue,...privateDeliveries].find(entry => entry.name === name);
    assert.ok(entry, 'Unknown event '+name);
    const records = await this.events(entry.owner, name, aggregate);
    assert.equal(records.length, 1, 'The scenario must identify exactly one original fact');
    const decoded = eventType.toObject(eventType.decode(records[0]!.body), {longs: String});
    assert.equal(decoded.visibility, entry.visibility);
    const id = randomUUID();
    decoded.id = id;
    const body = entry.visibility === 'domain' ? freshIdentity(records[0]!.body, id) :
      Buffer.from(eventType.encode(eventType.fromObject(decoded)).finish());
    const connection = await amqp.connect({
      hostname: '127.0.0.1', port: Number(this.values.AMQP_PORT), vhost: 'reference',
      username: 'cafe_'+entry.owner, password: this.values[entry.owner.toUpperCase()+'_BROKER_PASSWORD'],
    });
    try {
      const channel = await connection.createConfirmChannel();
      let returned = false;
      channel.on('return', () => { returned = true; });
      let timer: ReturnType<typeof setTimeout> | undefined;
      try {
        await Promise.race([
          new Promise<void>((resolve, reject) => channel.publish('cafe.events', entry.visibility+'.'+name, body,
            {persistent: true, mandatory: true, contentType: 'application/x-protobuf', messageId: id, type: name,
              appId: entry.owner, correlationId: decoded.correlationId,
              headers: {'contract-version': {'!': 'int32', value: 1}}},
            error => error ? reject(error) : resolve())),
          new Promise<never>((_, reject) => { timer = setTimeout(() => reject(new Error('Broker confirmation timed out')), 5_000); }),
        ]);
        assert.equal(returned, false, 'Mandatory event was returned without a route');
      } finally { clearTimeout(timer); }
    } finally { await connection.close(); }
    return {id, consumer: entry.consumer};
  }
}
