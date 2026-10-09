import {test} from 'node:test';
import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {setTimeout} from 'node:timers/promises';
import {Pool} from 'pg';
import amqp from 'amqplib';
import {createLoyaltyContainer} from '../apps/composition/loyalty.js';
import {PostgresDurableCommandOutbox, commandPublication} from './internal-commands.js';
import schema from '../contexts/loyalty/adaptors/messaging/generated/internal_commands.json' with {type: 'json'};
import {claim, finish} from './dispatch.js';
import {consume, confirmed, relay} from './broker.js';
import {processWorkers} from './diagnostics.js';
import {derivedId} from '../foundation/identity.js';
import type {Metadata} from '../foundation/application.js';

async function until(check: () => Promise<boolean>) {
  const deadline = Date.now()+12000;
  while (!await check()) { if (Date.now() > deadline) assert.fail('Command recovery did not converge'); await setTimeout(30); }
}
test('durable hand-off, rollback, lease recovery, bounded command retry and replay preserve one transition', async () => {
  assert.match(process.env.CAFE_DISPOSABLE_PROJECT ?? '', /^cafe-reference-test-/);
  const container = createLoyaltyContainer(process.env.LOYALTY_DATABASE_URL!);
  const c = container.cradle, db = c.database, codec = c.creditCodec;
  const admin = new Pool({connectionString: process.env.DATABASE_ADMIN_URL!.replace(/\/[^/?]+(?=\?|$)/, '/cafe_loyalty')});
  const source = randomUUID(), target = randomUUID();
  const command = {orderId: source, customerId: target};
  const m: Metadata = {id: derivedId(codec.consumer, source), target, name: codec.consumer, correlation: randomUUID(),
    causation: source, consumer: codec.consumer, sourceId: source, sourceHash: 'a'.repeat(64), input: command};
  const outbox = new PostgresDurableCommandOutbox(db, codec);
  const controller = new AbortController(), workers: Promise<void>[] = [];
  const connection = await amqp.connect(process.env.BROKER_ADMIN_URL!);
  const channel = await connection.createConfirmChannel();
  const queue = 'ref.'+codec.consumer+'.command';
  let attempts = 0, completed = 0, unavailable = true;
  try {
    await db.verify();
    // Force a failure after the immutable insert and before dispatch acceptance.
    await admin.query(`ALTER TABLE cafe.internal_command_dispatches ADD CONSTRAINT reject_handoff_probe CHECK(event_id <> '${m.id}') NOT VALID`);
    await assert.rejects(outbox.enqueue(m, command));
    assert.equal((await db.pool.query('SELECT count(*)::int AS n FROM cafe.internal_commands WHERE id=$1', [m.id])).rows[0].n, 0);
    await admin.query('ALTER TABLE cafe.internal_command_dispatches DROP CONSTRAINT reject_handoff_probe');
    await outbox.enqueue(m, command);
    await outbox.enqueue(m, command);
    await assert.rejects(outbox.enqueue({...m, sourceHash: 'b'.repeat(64)}, command));
    assert.equal(await c.accountQueries.get(target), undefined, 'Event acceptance mutated an aggregate');
    const saved = (await db.pool.query('SELECT body FROM cafe.internal_commands WHERE id=$1', [m.id])).rows[0].body as Buffer;
    const abandoned = await claim(db, 'commands'); assert.ok(abandoned); assert.equal(abandoned.id, m.id);
    await admin.query("UPDATE cafe.internal_command_dispatches SET lease_until=clock_timestamp()-interval '1 second' WHERE event_id=$1", [m.id]);
    const recovered = await claim(db, 'commands'); assert.ok(recovered);
    assert.deepEqual(recovered.body, saved);
    await finish(db, abandoned, 'commands');
    assert.equal((await db.pool.query('SELECT completed_at FROM cafe.internal_command_dispatches WHERE event_id=$1', [m.id])).rows[0].completed_at, null);
    await finish(db, recovered, 'commands', 'recover');
    workers.push(consume(process.env.LOYALTY_BROKER_URL!, {owner: 'loyalty', consumer: codec.consumer,
      event: 'collection.order-collected', target: () => target, decodeCommand: body => codec.decode(body),
      handle: async (metadata, payload) => {
        attempts++;
        if (unavailable) throw new Error('injected infrastructure outage');
        const outcome = await c.credit.execute(metadata, payload as typeof command);
        completed++; return outcome;
      }}, controller.signal));
    workers.push(relay(db, process.env.LOYALTY_BROKER_URL!, controller.signal, commandPublication(schema, 'loyalty')));
    let quarantined: amqp.GetMessage | false = false;
    await until(async () => { quarantined = await channel.get(queue+'.dead', {noAck: false}); return Boolean(quarantined); });
    assert.equal(attempts, 4);
    const failure = processWorkers()['loyalty/'+codec.consumer]?.lastFailure;
    assert.equal(failure?.event, m.id);
    assert.equal(failure?.correlation, m.correlation);
    const dead = quarantined as unknown as amqp.GetMessage;
    assert.deepEqual(dead.content, saved);
    assert.equal(await c.accountQueries.get(target), undefined);
    unavailable = false;
    const properties = {...dead.properties, headers: {'contract-version': 1}};
    await confirmed(channel, 'ref.loyalty.delivery', queue, dead.content, properties);
    channel.ack(dead);
    await until(async () => Boolean(await c.accountQueries.get(target)));
    // Lost ACK / duplicate publication reuses the committed command and source receipts.
    await confirmed(channel, 'ref.loyalty.delivery', queue, saved, properties);
    await until(async () => completed >= 2);
    const loaded = await c.accountQueries.get(target);
    assert.equal(loaded?.version, 1); assert.equal(loaded?.state.collections, 1);
    assert.equal((await db.pool.query('SELECT count(*)::int AS n FROM cafe.consumer_receipts WHERE consumer=$1 AND event_id=$2', [codec.consumer, source])).rows[0].n, 1);
    await assert.rejects(db.pool.query('UPDATE cafe.internal_commands SET body=$2 WHERE id=$1', [m.id, saved]));
  } finally {
    controller.abort(); await Promise.allSettled(workers);
    await admin.query('ALTER TABLE cafe.internal_command_dispatches DROP CONSTRAINT IF EXISTS reject_handoff_probe');
    await connection.close(); await container.dispose(); await admin.end();
  }
});
