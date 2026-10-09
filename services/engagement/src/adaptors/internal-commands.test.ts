import {readFileSync} from 'node:fs';
import {createCommunicationContainer} from '../apps/composition/communication.js';
import type {InternalCommandCodec} from './internal-commands.js';
import {test} from 'node:test';
import assert from 'node:assert/strict';
import {createLoyaltyContainer} from '../apps/composition/loyalty.js';
import {derivedId} from '../foundation/identity.js';
import type {Metadata} from '../foundation/application.js';

const id = '00000000-0000-4000-8000-000000000001';
const source = '00000000-0000-4000-8000-000000000002';
test('private command bytes round-trip typed intent and original receipt material deterministically', async () => {
  const container = createLoyaltyContainer('postgresql://unused@127.0.0.1:1/unused');
  const codec = container.resolve('creditCodec');
  const command = {orderId: source, customerId: id};
  const m: Metadata = {id: derivedId(codec.consumer, source), target: id, name: codec.consumer,
    correlation: id, causation: source, consumer: codec.consumer, sourceId: source, sourceHash: 'a'.repeat(64), input: command};
  try {
    const body = codec.encode(m, command);
    assert.deepEqual(codec.encode(m, command), body);
    assert.deepEqual(codec.decode(body), {metadata: m, payload: command});
    assert.throws(() => codec.encode({...m, consumer: 'communication.foreign'}, command));
    assert.throws(() => codec.encode({...m, id}, command));
    assert.throws(() => container.resolve('issueCodec').decode(body));
    assert.throws(() => codec.decode(body.subarray(0, 4)));
  } finally { await container.dispose(); }
});

test('every private command retains its historical wire fixture and subscription declaration', async () => {
  const loyalty = createLoyaltyContainer('postgresql://unused@127.0.0.1:1/unused');
  const communication = createCommunicationContainer('postgresql://unused@127.0.0.1:1/unused', 'http://unused', 'unused');
  function check<C extends object>(codec: InternalCommandCodec<C>, command: C) {
    const metadata: Metadata = {id: derivedId(codec.consumer, source), target: id, name: codec.consumer,
      correlation: id, causation: source, consumer: codec.consumer, sourceId: source, sourceHash: 'a'.repeat(64), input: {original: 'receipt'}};
    const saved = Buffer.from(readFileSync(new URL(`../contexts/${codec.owner}/adaptors/messaging/fixtures/${codec.consumer}.command.hex`, import.meta.url), 'utf8').trim(), 'hex');
    assert.deepEqual(codec.encode(metadata, command), saved);
    assert.deepEqual(codec.decode(saved), {metadata, payload: command});
  }
  try {
    check(loyalty.cradle.creditCodec, {orderId: source, customerId: id});
    check(loyalty.cradle.issueCodec, {grantId: source, accountId: id, benefit: 'coffee', validDays: 7});
    check(communication.cradle.pickupCodec, {recipient: id, subject: 'Ready', body: 'Collect with ABC123'});
    check(communication.cradle.rewardCodec, {recipient: id, subject: 'Reward', body: 'One coffee'});
    check(communication.cradle.deliveryCodec, {notificationId: id});
    for (const [owner, codecs] of [
      ['loyalty', [loyalty.cradle.creditCodec, loyalty.cradle.issueCodec]],
      ['communication', [communication.cradle.pickupCodec, communication.cradle.rewardCodec, communication.cradle.deliveryCodec]],
    ] as const) {
      const declared = JSON.parse(readFileSync(new URL(`../contexts/${owner}/adaptors/messaging/subscriptions.json`, import.meta.url), 'utf8')) as {consumer: string; command: string}[];
      assert.deepEqual(codecs.map(c => ({consumer: c.consumer, command: c.command})), declared.map(({consumer, command}) => ({consumer, command})));
    }
  } finally { await loyalty.dispose(); await communication.dispose(); }
});
